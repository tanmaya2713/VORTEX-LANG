package llvmir

import (
	"strconv"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"github.com/vortex-lang/vortex/src/ast"
	"github.com/vortex-lang/vortex/src/common"
	vtypes "github.com/vortex-lang/vortex/src/types"
)

func (c *Codegen) codegenExpr(expr ast.Expr) value.Value {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.NumberLit:
		return c.codegenNumber(e)
	case *ast.StringLit:
		return c.codegenStringLit(e)
	case *ast.BoolLit:
		if e.Value {
			return constant.NewInt(lltypes.I1, 1)
		}
		return constant.NewInt(lltypes.I1, 0)
	case *ast.Ident:
		return c.codegenIdent(e)
	case *ast.BinaryExpr:
		return c.codegenBinary(e)
	case *ast.UnaryExpr:
		return c.codegenUnary(e)
	case *ast.CallExpr:
		return c.codegenCall(e)
	case *ast.IndexExpr:
		return c.codegenIndex(e)
	case *ast.ArrayLit:
		return c.codegenArrayLit(e)
	case *ast.MemberExpr:
		return c.codegenMember(e)
	default:
		return nil
	}
}

func (c *Codegen) exprType(expr ast.Expr) vtypes.Type {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.Ident:
		if vi, ok := c.sc.get(e.Name); ok {
			return vi.typ
		}
		return vtypes.TypeVoid
	case *ast.NumberLit:
		if hasDecimal(e.Value) {
			return vtypes.TypeF64
		}
		return vtypes.TypeI32
	case *ast.StringLit:
		return vtypes.TypeString
	case *ast.BoolLit:
		return vtypes.TypeBool
	case *ast.BinaryExpr:
		return c.exprTypeBinary(e)
	case *ast.UnaryExpr:
		return c.exprType(e.Right)
	case *ast.CallExpr:
		return c.exprTypeCall(e)
	case *ast.IndexExpr:
		return c.exprTypeIndex(e)
	case *ast.ArrayLit:
		return c.exprTypeArrayLit(e)
	case *ast.MemberExpr:
		return c.exprTypeMember(e)
	default:
		return vtypes.TypeVoid
	}
}

func (c *Codegen) codegenNumber(e *ast.NumberLit) value.Value {
	if hasDecimal(e.Value) {
		val, _ := strconv.ParseFloat(e.Value, 64)
		return constant.NewFloat(lltypes.Double, val)
	}
	val, _ := strconv.ParseInt(e.Value, 10, 32)
	return constant.NewInt(lltypes.I32, val)
}

func (c *Codegen) codegenIdent(e *ast.Ident) value.Value {
	if vi, ok := c.sc.get(e.Name); ok {
		load := ir.NewLoad(vi.alloca.Type().(*lltypes.PointerType).ElemType, vi.alloca)
		c.currentBlock.Insts = append(c.currentBlock.Insts, load)
		return load
	}
	return nil
}

func (c *Codegen) codegenStringLit(e *ast.StringLit) value.Value {
	return c.codegenStrConstant(e.Value)
}

func (c *Codegen) codegenStrConstant(content string) value.Value {
	g := c.module.NewGlobalDef(".str", constant.NewCharArrayFromString(content+"\x00"))
	g.Immutable = true
	zero := constant.NewInt(lltypes.I32, 0)
	ptrType := g.Type().(*lltypes.PointerType)
	gep := ir.NewGetElementPtr(ptrType.ElemType, g, zero, zero)
	c.currentBlock.Insts = append(c.currentBlock.Insts, gep)
	return gep
}

func (c *Codegen) codegenBinary(e *ast.BinaryExpr) value.Value {
	left := c.codegenExpr(e.Left)
	right := c.codegenExpr(e.Right)
	if left == nil || right == nil {
		return nil
	}

	leftType := c.exprType(e.Left)
	rightType := c.exprType(e.Right)
	isFloat := leftType != nil && (leftType.Kind() == common.TypeF32 || leftType.Kind() == common.TypeF64)

	switch e.Op {
	case "+":
		if leftType.Kind() == common.TypeTensor {
			return c.codegenTensorAdd(left, right, leftType, rightType)
		}
		if isFloat {
			inst := ir.NewFAdd(left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewAdd(left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "-":
		if isFloat {
			inst := ir.NewFSub(left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewSub(left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "*":
		if leftType.Kind() == common.TypeTensor {
			return c.codegenTensorMatmul(left, right, leftType, rightType)
		}
		if isFloat {
			inst := ir.NewFMul(left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewMul(left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "/":
		if isFloat {
			inst := ir.NewFDiv(left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewSDiv(left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "==":
		if isFloat {
			inst := ir.NewFCmp(enum.FPredOEQ, left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewICmp(enum.IPredEQ, left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "!=":
		if isFloat {
			inst := ir.NewFCmp(enum.FPredONE, left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewICmp(enum.IPredNE, left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "<":
		if isFloat {
			inst := ir.NewFCmp(enum.FPredOLT, left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewICmp(enum.IPredSLT, left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case ">":
		if isFloat {
			inst := ir.NewFCmp(enum.FPredOGT, left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewICmp(enum.IPredSGT, left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "<=":
		if isFloat {
			inst := ir.NewFCmp(enum.FPredOLE, left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewICmp(enum.IPredSLE, left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case ">=":
		if isFloat {
			inst := ir.NewFCmp(enum.FPredOGE, left, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		inst := ir.NewICmp(enum.IPredSGE, left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "&&":
		inst := ir.NewAnd(left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "||":
		inst := ir.NewOr(left, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	}
	return nil
}

func (c *Codegen) exprTypeBinary(e *ast.BinaryExpr) vtypes.Type {
	leftType := c.exprType(e.Left)
	if leftType == nil {
		return vtypes.TypeVoid
	}

	if (e.Op == "+" || e.Op == "*") && leftType.Kind() == common.TypeTensor {
		if rightType, ok := c.exprType(e.Right).(*vtypes.TensorType); ok {
			if leftTensor, ok := leftType.(*vtypes.TensorType); ok && len(leftTensor.Shape) == 2 && len(rightType.Shape) == 2 {
				return &vtypes.TensorType{
					ElemType: leftTensor.ElemType,
					Shape:    []int64{leftTensor.Shape[0], rightType.Shape[1]},
				}
			}
		}
		return vtypes.TypeVoid
	}

	rightType := c.exprType(e.Right)
	if rightType == nil {
		return vtypes.TypeVoid
	}

	switch e.Op {
	case "==", "!=", "<", ">", "<=", ">=":
		return vtypes.TypeBool
	}
	if leftType.Kind() == common.TypeF64 || rightType.Kind() == common.TypeF64 {
		return vtypes.TypeF64
	}
	return leftType
}

func (c *Codegen) codegenUnary(e *ast.UnaryExpr) value.Value {
	right := c.codegenExpr(e.Right)
	if right == nil {
		return nil
	}
	switch e.Op {
	case "-":
		rightType := c.exprType(e.Right)
		if rightType != nil && rightType.Kind() == common.TypeF64 {
			zero := constant.NewFloat(lltypes.Double, 0.0)
			inst := ir.NewFSub(zero, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		if rightType != nil && rightType.Kind() == common.TypeF32 {
			zero := constant.NewFloat(lltypes.Float, 0.0)
			inst := ir.NewFSub(zero, right)
			c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
			return inst
		}
		zero := constant.NewInt(lltypes.I32, 0)
		inst := ir.NewSub(zero, right)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	case "!":
		trueVal := constant.NewInt(lltypes.I1, 1)
		inst := ir.NewXor(right, trueVal)
		c.currentBlock.Insts = append(c.currentBlock.Insts, inst)
		return inst
	}
	return nil
}

func (c *Codegen) codegenCall(e *ast.CallExpr) value.Value {
	if ident, ok := e.Fn.(*ast.Ident); ok {
		switch ident.Name {
		case "relu":
			return c.codegenRelu(ident, e.Args)
		case "sigmoid":
			return c.codegenSigmoid(ident, e.Args)
		}
		if fn, ok := c.funcMap[ident.Name]; ok {
			args := make([]value.Value, len(e.Args))
			for i, arg := range e.Args {
				args[i] = c.codegenExpr(arg)
				if args[i] == nil {
					return nil
				}
			}
			if ident.Name == "print" || ident.Name == "println" {
				for _, arg := range args {
					c.emitCallVoid(c.rt.vortexPrintI32, arg)
				}
				return nil
			}
			if fn.Sig.RetType == lltypes.Void {
				c.emitCallVoid(fn, args...)
				return nil
			}
			return c.emitCall(fn, args...)
		}
	}
	return nil
}

func (c *Codegen) codegenIndex(e *ast.IndexExpr) value.Value {
	target := c.codegenExpr(e.Target)
	idx := c.codegenExpr(e.Index)
	if target == nil || idx == nil {
		return nil
	}

	targetType := c.exprType(e.Target)
	if targetType != nil && targetType.Kind() == common.TypeTensor {
		tensorPtrType := target.Type()
		tmpAlloca := ir.NewAlloca(tensorPtrType)
		c.currentBlock.Insts = append(c.currentBlock.Insts, tmpAlloca)
		tmpStore := ir.NewStore(target, tmpAlloca)
		c.currentBlock.Insts = append(c.currentBlock.Insts, tmpStore)

		zero := constant.NewInt(lltypes.I32, 0)
		tensorPtrVal := ir.NewLoad(tensorPtrType, tmpAlloca)
		c.currentBlock.Insts = append(c.currentBlock.Insts, tensorPtrVal)

		dataPtr := ir.NewGetElementPtr(getTensorStructType(), tensorPtrVal, zero, constant.NewInt(lltypes.I32, 2))
		c.currentBlock.Insts = append(c.currentBlock.Insts, dataPtr)

		dataLoad := ir.NewLoad(lltypes.NewPointer(lltypes.Float), dataPtr)
		c.currentBlock.Insts = append(c.currentBlock.Insts, dataLoad)

		elemPtr := ir.NewGetElementPtr(lltypes.Float, dataLoad, idx)
		c.currentBlock.Insts = append(c.currentBlock.Insts, elemPtr)

		elemLoad := ir.NewLoad(lltypes.Float, elemPtr)
		c.currentBlock.Insts = append(c.currentBlock.Insts, elemLoad)
		return elemLoad
	}

	arrLLVMType := target.Type()
	tmpAlloca := ir.NewAlloca(arrLLVMType)
	c.currentBlock.Insts = append(c.currentBlock.Insts, tmpAlloca)
	tmpStore := ir.NewStore(target, tmpAlloca)
	c.currentBlock.Insts = append(c.currentBlock.Insts, tmpStore)

	zero := constant.NewInt(lltypes.I32, 0)
	gep := ir.NewGetElementPtr(arrLLVMType, tmpAlloca, zero, idx)
	c.currentBlock.Insts = append(c.currentBlock.Insts, gep)

	elemLoad := ir.NewLoad(gep.Type().(*lltypes.PointerType).ElemType, gep)
	c.currentBlock.Insts = append(c.currentBlock.Insts, elemLoad)
	return elemLoad
}

func (c *Codegen) codegenArrayLit(e *ast.ArrayLit) value.Value {
	if len(e.Elems) == 0 {
		return nil
	}

	elemType := c.exprType(e.Elems[0])
	if elemType == nil {
		return nil
	}
	llvmElemType := toLLVMType(elemType)
	arrType := lltypes.NewArray(uint64(len(e.Elems)), llvmElemType)

	alloca := ir.NewAlloca(arrType)
	c.currentBlock.Insts = append(c.currentBlock.Insts, alloca)

	zero := constant.NewInt(lltypes.I32, 0)
	for i, elem := range e.Elems {
		v := c.codegenExpr(elem)
		if v == nil {
			return nil
		}
		gep := ir.NewGetElementPtr(arrType, alloca, zero, constant.NewInt(lltypes.I32, int64(i)))
		c.currentBlock.Insts = append(c.currentBlock.Insts, gep)
		store := ir.NewStore(v, gep)
		c.currentBlock.Insts = append(c.currentBlock.Insts, store)
	}

	load := ir.NewLoad(arrType, alloca)
	c.currentBlock.Insts = append(c.currentBlock.Insts, load)
	return load
}

func (c *Codegen) codegenMember(e *ast.MemberExpr) value.Value {
	target := c.codegenExpr(e.Target)
	if target == nil {
		return nil
	}
	_ = target
	return nil
}

func (c *Codegen) codegenTensorMatmul(left, right value.Value, leftType, rightType vtypes.Type) value.Value {
	leftTensor, ok := leftType.(*vtypes.TensorType)
	if !ok || len(leftTensor.Shape) != 2 {
		c.addError(common.Position{}, "tensor matmul requires 2D left operand")
		return nil
	}
	rightTensor, ok := rightType.(*vtypes.TensorType)
	if !ok || len(rightTensor.Shape) != 2 {
		c.addError(common.Position{}, "tensor matmul requires 2D right operand")
		return nil
	}

	M := leftTensor.Shape[0]
	N := rightTensor.Shape[1]

	tensorPtrType := toLLVMType(leftType)

	tmpOutAlloca := ir.NewAlloca(tensorPtrType)
	c.currentBlock.Insts = append(c.currentBlock.Insts, tmpOutAlloca)

	zero := constant.NewInt(lltypes.I32, 0)
	shapeArrType := lltypes.NewArray(2, lltypes.I32)
	shapeAlloca := ir.NewAlloca(shapeArrType)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeAlloca)

	shapeInit := constant.NewZeroInitializer(shapeArrType)
	shapeStore := ir.NewStore(shapeInit, shapeAlloca)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeStore)

	mVal := constant.NewInt(lltypes.I32, M)
	gep0 := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, zero)
	c.currentBlock.Insts = append(c.currentBlock.Insts, gep0)
	c.currentBlock.Insts = append(c.currentBlock.Insts, ir.NewStore(mVal, gep0))

	nVal := constant.NewInt(lltypes.I32, N)
	gep1 := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, constant.NewInt(lltypes.I32, 1))
	c.currentBlock.Insts = append(c.currentBlock.Insts, gep1)
	c.currentBlock.Insts = append(c.currentBlock.Insts, ir.NewStore(nVal, gep1))

	shapePtr := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, zero)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapePtr)

	ndimVal := constant.NewInt(lltypes.I32, 2)

	nullPtr := constant.NewZeroInitializer(lltypes.NewPointer(lltypes.Float))

	outTensor := c.emitCall(c.rt.vortexTensorCreate, shapePtr, ndimVal, nullPtr)

	c.emitCallVoid(c.rt.vortexTensorMatmul, left, right, outTensor)

	return outTensor
}

func (c *Codegen) codegenTensorAdd(left, right value.Value, leftType, rightType vtypes.Type) value.Value {
	leftTensor, ok := leftType.(*vtypes.TensorType)
	if !ok {
		c.addError(common.Position{}, "tensor add requires both operands to be tensors")
		return nil
	}

	zero := constant.NewInt(lltypes.I32, 0)

	ndimVal := constant.NewInt(lltypes.I32, int64(len(leftTensor.Shape)))
	shapeArrType := lltypes.NewArray(uint64(len(leftTensor.Shape)), lltypes.I32)
	shapeAlloca := ir.NewAlloca(shapeArrType)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeAlloca)

	shapeInit := constant.NewZeroInitializer(shapeArrType)
	shapeStore := ir.NewStore(shapeInit, shapeAlloca)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeStore)

	for i, dim := range leftTensor.Shape {
		dimVal := constant.NewInt(lltypes.I32, dim)
		gep := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, constant.NewInt(lltypes.I32, int64(i)))
		c.currentBlock.Insts = append(c.currentBlock.Insts, gep)
		c.currentBlock.Insts = append(c.currentBlock.Insts, ir.NewStore(dimVal, gep))
	}

	shapePtr := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, zero)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapePtr)

	nullPtr := constant.NewZeroInitializer(lltypes.NewPointer(lltypes.Float))
	outTensor := c.emitCall(c.rt.vortexTensorCreate, shapePtr, ndimVal, nullPtr)

	c.emitCallVoid(c.rt.vortexTensorAdd, left, right, outTensor)

	return outTensor
}

func (c *Codegen) codegenRelu(ident *ast.Ident, args []ast.Expr) value.Value {
	if len(args) != 1 {
		c.addError(ident.Pos(), "relu requires exactly 1 argument")
		return nil
	}
	inputVal := c.codegenExpr(args[0])
	inputType := c.exprType(args[0])
	if inputType == nil || inputType.Kind() != common.TypeTensor {
		c.addError(ident.Pos(), "relu requires a tensor argument")
		return nil
	}

	tensorType := inputType.(*vtypes.TensorType)
	zero := constant.NewInt(lltypes.I32, 0)

	ndimVal := constant.NewInt(lltypes.I32, int64(len(tensorType.Shape)))
	shapeArrType := lltypes.NewArray(uint64(len(tensorType.Shape)), lltypes.I32)
	shapeAlloca := ir.NewAlloca(shapeArrType)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeAlloca)

	shapeInit := constant.NewZeroInitializer(shapeArrType)
	shapeStore := ir.NewStore(shapeInit, shapeAlloca)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeStore)

	for i, dim := range tensorType.Shape {
		dimVal := constant.NewInt(lltypes.I32, dim)
		gep := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, constant.NewInt(lltypes.I32, int64(i)))
		c.currentBlock.Insts = append(c.currentBlock.Insts, gep)
		c.currentBlock.Insts = append(c.currentBlock.Insts, ir.NewStore(dimVal, gep))
	}

	shapePtr := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, zero)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapePtr)

	nullPtr := constant.NewZeroInitializer(lltypes.NewPointer(lltypes.Float))
	outTensor := c.emitCall(c.rt.vortexTensorCreate, shapePtr, ndimVal, nullPtr)

	c.emitCallVoid(c.rt.vortexTensorRelu, inputVal, outTensor)

	return outTensor
}

func (c *Codegen) codegenSigmoid(ident *ast.Ident, args []ast.Expr) value.Value {
	if len(args) != 1 {
		c.addError(ident.Pos(), "sigmoid requires exactly 1 argument")
		return nil
	}
	inputVal := c.codegenExpr(args[0])
	inputType := c.exprType(args[0])
	if inputType == nil || inputType.Kind() != common.TypeTensor {
		c.addError(ident.Pos(), "sigmoid requires a tensor argument")
		return nil
	}

	tensorType := inputType.(*vtypes.TensorType)
	zero := constant.NewInt(lltypes.I32, 0)

	ndimVal := constant.NewInt(lltypes.I32, int64(len(tensorType.Shape)))
	shapeArrType := lltypes.NewArray(uint64(len(tensorType.Shape)), lltypes.I32)
	shapeAlloca := ir.NewAlloca(shapeArrType)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeAlloca)

	shapeInit := constant.NewZeroInitializer(shapeArrType)
	shapeStore := ir.NewStore(shapeInit, shapeAlloca)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapeStore)

	for i, dim := range tensorType.Shape {
		dimVal := constant.NewInt(lltypes.I32, dim)
		gep := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, constant.NewInt(lltypes.I32, int64(i)))
		c.currentBlock.Insts = append(c.currentBlock.Insts, gep)
		c.currentBlock.Insts = append(c.currentBlock.Insts, ir.NewStore(dimVal, gep))
	}

	shapePtr := ir.NewGetElementPtr(shapeArrType, shapeAlloca, zero, zero)
	c.currentBlock.Insts = append(c.currentBlock.Insts, shapePtr)

	nullPtr := constant.NewZeroInitializer(lltypes.NewPointer(lltypes.Float))
	outTensor := c.emitCall(c.rt.vortexTensorCreate, shapePtr, ndimVal, nullPtr)

	c.emitCallVoid(c.rt.vortexTensorSigmoid, inputVal, outTensor)

	return outTensor
}

func (c *Codegen) emitCall(fn value.Value, args ...value.Value) value.Value {
	call := ir.NewCall(fn, args...)
	c.currentBlock.Insts = append(c.currentBlock.Insts, call)
	return call
}

func (c *Codegen) emitCallVoid(fn value.Value, args ...value.Value) {
	call := ir.NewCall(fn, args...)
	c.currentBlock.Insts = append(c.currentBlock.Insts, call)
}

func (c *Codegen) exprTypeCall(e *ast.CallExpr) vtypes.Type {
	if ident, ok := e.Fn.(*ast.Ident); ok {
		switch ident.Name {
		case "relu", "sigmoid":
			if len(e.Args) == 1 {
				return c.exprType(e.Args[0])
			}
		}
		if ft, ok := c.funcTypes[ident.Name]; ok {
			return ft.(*vtypes.FnType).ReturnType
		}
	}
	return vtypes.TypeVoid
}

func (c *Codegen) exprTypeIndex(e *ast.IndexExpr) vtypes.Type {
	targetType := c.exprType(e.Target)
	if targetType == nil {
		return vtypes.TypeVoid
	}
	if tensorType, ok := targetType.(*vtypes.TensorType); ok {
		return tensorType.ElemType
	}
	if arrType, ok := targetType.(*vtypes.ArrayType); ok {
		return arrType.ElemType
	}
	return vtypes.TypeVoid
}

func (c *Codegen) exprTypeArrayLit(e *ast.ArrayLit) vtypes.Type {
	if len(e.Elems) == 0 {
		return vtypes.TypeVoid
	}
	elemType := c.exprType(e.Elems[0])
	if elemType == nil {
		return vtypes.TypeVoid
	}
	return &vtypes.ArrayType{ElemType: elemType, Len: int64(len(e.Elems))}
}

func (c *Codegen) exprTypeMember(e *ast.MemberExpr) vtypes.Type {
	targetType := c.exprType(e.Target)
	if targetType == nil {
		return vtypes.TypeVoid
	}
	if st, ok := targetType.(*vtypes.StructType); ok {
		if ft, ok := st.Fields[e.Name.String()]; ok {
			return ft
		}
	}
	return vtypes.TypeVoid
}
