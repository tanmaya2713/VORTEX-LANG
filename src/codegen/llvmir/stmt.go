package llvmir

import (
	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"github.com/vortex-lang/vortex/src/ast"
	"github.com/vortex-lang/vortex/src/common"
	vtypes "github.com/vortex-lang/vortex/src/types"
)

func (c *Codegen) codegenStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		c.codegenLet(s)
	case *ast.FnDef:
	case *ast.IfStmt:
		c.codegenIf(s)
	case *ast.WhileStmt:
		c.codegenWhile(s)
	case *ast.ForStmt:
		c.codegenFor(s)
	case *ast.ReturnStmt:
		c.codegenReturn(s)
	case *ast.BreakStmt:
		c.codegenBreak()
	case *ast.ContinueStmt:
		c.codegenContinue()
	case *ast.BlockStmt:
		c.codegenBlock(s)
	case *ast.PrintStmt:
		c.codegenPrint(s)
	case *ast.AssignExpr:
		c.codegenAssign(s)
	case *ast.ExprStmt:
		if assign, ok := s.E.(*ast.AssignExpr); ok {
			c.codegenAssign(assign)
		} else {
			c.codegenExpr(s.E)
		}
	case *ast.ImportStmt:
	case *ast.AssertStmt:
		c.codegenAssert(s)
	case *ast.StructDef:
	case *ast.ModelDef:
		c.codegenModelDef(s)
	case *ast.TrainStmt:
		c.codegenTrain(s)
	}
}

func (c *Codegen) codegenLet(s *ast.LetStmt) {
	var varType vtypes.Type
	if s.Type != nil {
		varType = c.resolveTypeExpr(s.Type)
	} else if s.Value != nil {
		varType = c.exprType(s.Value)
	} else {
		return
	}

	if tensorType, ok := varType.(*vtypes.TensorType); ok && s.Value != nil {
		if _, ok := s.Value.(*ast.ArrayLit); ok {
			c.codegenTensorLet(s, tensorType)
			return
		}
	}

	var initVal value.Value
	if s.Value != nil {
		initVal = c.codegenExpr(s.Value)
	}

	llvmTyp := toLLVMType(varType)
	alloca := ir.NewAlloca(llvmTyp)
	alloca.SetName(s.Name.String())
	c.currentBlock.Insts = append(c.currentBlock.Insts, alloca)

	if initVal != nil {
		store := ir.NewStore(initVal, alloca)
		c.currentBlock.Insts = append(c.currentBlock.Insts, store)
	} else {
		zeroInit := constant.NewZeroInitializer(llvmTyp)
		store := ir.NewStore(zeroInit, alloca)
		c.currentBlock.Insts = append(c.currentBlock.Insts, store)
	}

	c.sc.set(s.Name.String(), &varInfo{alloca: alloca, typ: varType})
}

func (c *Codegen) codegenTensorLet(s *ast.LetStmt, tensorType *vtypes.TensorType) {
	shapeVals := make([]value.Value, len(tensorType.Shape))
	totalElems := int64(1)
	for i, d := range tensorType.Shape {
		shapeVals[i] = constant.NewInt(lltypes.I32, d)
		totalElems *= d
	}

	ndim := len(tensorType.Shape)
	shapeArrType := lltypes.NewArray(uint64(ndim), lltypes.I32)
	shapeArr := constant.NewZeroInitializer(shapeArrType)
	shapeAlloca := ir.NewAlloca(shapeArrType)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeAlloca)
	shapeStore := ir.NewStore(shapeArr, shapeAlloca)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeStore)
	zero := constant.NewInt(lltypes.I32, 0)
	for i, sv := range shapeVals {
		gep := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, constant.NewInt(lltypes.I32, int64(i)))
		c.currentBlock.Insts = append(c.currentBlock.Insts, gep)
		elStore := ir.NewStore(sv, gep)
		c.currentBlock.Insts = append(c.currentBlock.Insts, elStore)
	}

	shapePtr := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, zero)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapePtr)

	arrVal := c.codegenExpr(s.Value)
	if arrVal == nil {
		return
	}

	arrAlloca := ir.NewAlloca(arrVal.Type())
	c.currentBlock.Insts = append(c.currentBlock.Insts, arrAlloca)
	arrStore := ir.NewStore(arrVal, arrAlloca)
	c.currentBlock.Insts = append(c.currentBlock.Insts, arrStore)

	dataPtr := ir.NewGetElementPtr(arrVal.Type(), arrAlloca, zero, zero)
	c.currentBlock.Insts = append(c.currentBlock.Insts, dataPtr)

	ndimVal := constant.NewInt(lltypes.I32, int64(ndim))
	result := c.emitCall(c.rt.vortexTensorCreate, shapePtr, ndimVal, dataPtr)

	tensorPtrType := toLLVMType(tensorType)
	alloca := ir.NewAlloca(tensorPtrType)
	alloca.SetName(s.Name.String())
	c.currentBlock.Insts = append(c.currentBlock.Insts, alloca)
	store := ir.NewStore(result, alloca)
	c.currentBlock.Insts = append(c.currentBlock.Insts, store)

	c.sc.set(s.Name.String(), &varInfo{alloca: alloca, typ: tensorType})
}

func (c *Codegen) codegenBlock(s *ast.BlockStmt) {
	c.sc.push()
	for _, stmt := range s.Stmts {
		if c.currentBlock != nil && c.currentBlock.Term != nil {
			break
		}
		c.codegenStmt(stmt)
	}
	c.sc.pop()
}

func (c *Codegen) codegenIf(s *ast.IfStmt) {
	cond := c.codegenExpr(s.Cond)

	thenBlock := c.freshBlock("if_then")
	elseBlock := c.freshBlock("if_else")
	mergeBlock := c.freshBlock("if_merge")

	c.emitTermCondBr(cond, thenBlock, elseBlock)

	c.setBlock(thenBlock)
	c.codegenBlock(s.Then)
	if c.currentBlock.Term == nil {
		c.emitTermBr(mergeBlock)
	}

	c.setBlock(elseBlock)
	if s.Else != nil {
		c.codegenStmt(s.Else)
	}
	if c.currentBlock.Term == nil {
		c.emitTermBr(mergeBlock)
	}

	c.setBlock(mergeBlock)
}

func (c *Codegen) codegenWhile(s *ast.WhileStmt) {
	headerBlock := c.freshBlock("while_header")
	bodyBlock := c.freshBlock("while_body")
	exitBlock := c.freshBlock("while_exit")

	c.emitTermBr(headerBlock)

	c.setBlock(headerBlock)
	cond := c.codegenExpr(s.Cond)
	c.emitTermCondBr(cond, bodyBlock, exitBlock)

	c.breakStack = append([]*ir.Block{exitBlock}, c.breakStack...)
	c.continueStack = append([]*ir.Block{headerBlock}, c.continueStack...)

	c.setBlock(bodyBlock)
	c.codegenBlock(s.Body)
	if c.currentBlock.Term == nil {
		c.emitTermBr(headerBlock)
	}

	c.breakStack = c.breakStack[1:]
	c.continueStack = c.continueStack[1:]

	c.setBlock(exitBlock)
}

func (c *Codegen) codegenFor(s *ast.ForStmt) {
	rangeVal := c.codegenExpr(s.Range)
	rangeType := c.exprType(s.Range)
	if rangeVal == nil {
		return
	}

	arrayType, ok := rangeType.(*vtypes.ArrayType)
	if !ok {
		return
	}
	elemType := arrayType.ElemType
	llvmElemType := toLLVMType(elemType)

	preheaderBlock := c.currentBlock
	headerBlock := c.freshBlock("for_header")
	bodyBlock := c.freshBlock("for_body")
	exitBlock := c.freshBlock("for_exit")

	idxAlloca := ir.NewAlloca(lltypes.I32)
	idxAlloca.SetName(s.Var.Name + "_idx")
	preheaderBlock.Insts = append(preheaderBlock.Insts, idxAlloca)
	zeroStore := ir.NewStore(constant.NewInt(lltypes.I32, 0), idxAlloca)
	preheaderBlock.Insts = append(preheaderBlock.Insts, zeroStore)

	lenVal := constant.NewInt(lltypes.I32, int64(arrayType.Len))

	preheaderBlock.Term = ir.NewBr(headerBlock)
	c.currentBlock = headerBlock
	c.currentFunc.Blocks = append(c.currentFunc.Blocks, headerBlock)

	idxLoad := ir.NewLoad(lltypes.I32, idxAlloca)
	headerBlock.Insts = append(headerBlock.Insts, idxLoad)
	cond := ir.NewICmp(enum.IPredSLT, idxLoad, lenVal)
	headerBlock.Insts = append(headerBlock.Insts, cond)
	headerBlock.Term = ir.NewCondBr(cond, bodyBlock, exitBlock)

	c.breakStack = append([]*ir.Block{exitBlock}, c.breakStack...)
	c.continueStack = append([]*ir.Block{headerBlock}, c.continueStack...)

	c.currentBlock = bodyBlock
	c.currentFunc.Blocks = append(c.currentFunc.Blocks, bodyBlock)
	c.sc.push()

	targetValType := rangeVal.Type()
	arrAlloca := ir.NewAlloca(targetValType)
	bodyBlock.Insts = append(bodyBlock.Insts, arrAlloca)
	arrStore := ir.NewStore(rangeVal, arrAlloca)
	bodyBlock.Insts = append(bodyBlock.Insts, arrStore)

	elemPtr := ir.NewGetElementPtr(targetValType, arrAlloca,
		constant.NewInt(lltypes.I32, 0),
		idxLoad,
	)
	bodyBlock.Insts = append(bodyBlock.Insts, elemPtr)
	elemLoad := ir.NewLoad(llvmElemType, elemPtr)
	bodyBlock.Insts = append(bodyBlock.Insts, elemLoad)

	elemAlloca := ir.NewAlloca(llvmElemType)
	elemAlloca.SetName(s.Var.Name)
	bodyBlock.Insts = append(bodyBlock.Insts, elemAlloca)
	elemStore := ir.NewStore(elemLoad, elemAlloca)
	bodyBlock.Insts = append(bodyBlock.Insts, elemStore)

	c.sc.set(s.Var.Name, &varInfo{alloca: elemAlloca, typ: elemType})

	c.codegenBlock(s.Body)

	c.sc.pop()

	if c.currentBlock.Term == nil {
		nextIdx := ir.NewAdd(idxLoad, constant.NewInt(lltypes.I32, 1))
		c.currentBlock.Insts = append(c.currentBlock.Insts, nextIdx)
		storeNext := ir.NewStore(nextIdx, idxAlloca)
		c.currentBlock.Insts = append(c.currentBlock.Insts, storeNext)
		c.currentBlock.Term = ir.NewBr(headerBlock)
	}

	c.breakStack = c.breakStack[1:]
	c.continueStack = c.continueStack[1:]

	c.currentBlock = exitBlock
	c.currentFunc.Blocks = append(c.currentFunc.Blocks, exitBlock)
}

func (c *Codegen) codegenReturn(s *ast.ReturnStmt) {
	if s.Value != nil {
		val := c.codegenExpr(s.Value)
		if val != nil {
			c.emitTermRet(val)
		}
	} else {
		c.emitTermRet(nil)
	}
}

func (c *Codegen) codegenBreak() {
	if len(c.breakStack) > 0 {
		c.emitTermBr(c.breakStack[0])
	}
}

func (c *Codegen) codegenContinue() {
	if len(c.continueStack) > 0 {
		c.emitTermBr(c.continueStack[0])
	}
}

func (c *Codegen) codegenPrint(s *ast.PrintStmt) {
	val := c.codegenExpr(s.Expr)
	if val == nil {
		return
	}
	valType := c.exprType(s.Expr)
	switch valType.Kind() {
	case common.TypeI32, common.TypeI64, common.TypeI8, common.TypeI16,
		common.TypeU8, common.TypeU16, common.TypeU32, common.TypeU64:
		c.emitCallVoid(c.rt.vortexPrintI32, val)
	case common.TypeF32, common.TypeF64:
		c.emitCallVoid(c.rt.vortexPrintF64, val)
	case common.TypeBool:
		c.emitCallVoid(c.rt.vortexPrintBool, val)
	case common.TypeString:
		c.emitCallVoid(c.rt.vortexPrintStr, val)
	}
	c.emitCallVoid(c.rt.vortexPrintNewline)
}

func (c *Codegen) codegenAssign(s *ast.AssignExpr) {
	val := c.codegenExpr(s.Right)
	if val == nil {
		return
	}

	switch target := s.Left.(type) {
	case *ast.Ident:
		if vi, ok := c.sc.get(target.Name); ok {
			store := ir.NewStore(val, vi.alloca)
			c.currentBlock.Insts = append(c.currentBlock.Insts, store)
		}
	case *ast.IndexExpr:
		arrVal := c.codegenExpr(target.Target)
		idx := c.codegenExpr(target.Index)
		if arrVal != nil && idx != nil {
			arrLLVMType := arrVal.Type()
			tmpAlloca := ir.NewAlloca(arrLLVMType)
			c.currentBlock.Insts = append(c.currentBlock.Insts, tmpAlloca)
			tmpStore := ir.NewStore(arrVal, tmpAlloca)
			c.currentBlock.Insts = append(c.currentBlock.Insts, tmpStore)

			zero := constant.NewInt(lltypes.I32, 0)
			gep := ir.NewGetElementPtr(arrLLVMType, tmpAlloca, zero, idx)
			c.currentBlock.Insts = append(c.currentBlock.Insts, gep)
			store := ir.NewStore(val, gep)
			c.currentBlock.Insts = append(c.currentBlock.Insts, store)

			storeBack := ir.NewLoad(arrLLVMType, tmpAlloca)
			c.currentBlock.Insts = append(c.currentBlock.Insts, storeBack)
			_ = storeBack
		}
	}
}

func (c *Codegen) codegenAssert(s *ast.AssertStmt) {
	cond := c.codegenExpr(s.Cond)
	if cond == nil {
		return
	}

	failBlock := c.freshBlock("assert_fail")
	contBlock := c.freshBlock("assert_cont")
	c.emitTermCondBr(cond, contBlock, failBlock)

	c.setBlock(failBlock)
	if s.Msg != nil {
		msgVal := c.codegenExpr(s.Msg)
		if msgVal != nil {
			fmtStr := c.codegenStrConstant("assertion failed: %s\n")
			c.emitCallVoid(c.rt.printf, fmtStr, msgVal)
		}
	} else {
		fmtStr := c.codegenStrConstant("assertion failed\n")
		c.emitCallVoid(c.rt.printf, fmtStr)
	}
	c.emitTermRet(constant.NewInt(lltypes.I32, 1))

	c.setBlock(contBlock)
}

func (c *Codegen) codegenModelDef(s *ast.ModelDef) {
	c.emitCallVoid(c.rt.vortexInit)
	_ = s
}

func (c *Codegen) codegenTrain(s *ast.TrainStmt) {
	_ = s
}
