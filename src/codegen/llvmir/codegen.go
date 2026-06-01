package llvmir

import (
	"fmt"
	"strings"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
	"github.com/vortex-lang/vortex/src/ast"
	"github.com/vortex-lang/vortex/src/common"
	vtypes "github.com/vortex-lang/vortex/src/types"
)

type varInfo struct {
	alloca value.Value
	typ    vtypes.Type
}

type scope struct {
	vars []map[string]*varInfo
}

func (s *scope) push() {
	s.vars = append(s.vars, make(map[string]*varInfo))
}

func (s *scope) pop() {
	s.vars = s.vars[:len(s.vars)-1]
}

func (s *scope) set(name string, vi *varInfo) {
	s.vars[len(s.vars)-1][name] = vi
}

func (s *scope) get(name string) (*varInfo, bool) {
	for i := len(s.vars) - 1; i >= 0; i-- {
		if vi, ok := s.vars[i][name]; ok {
			return vi, ok
		}
	}
	return nil, false
}

type Codegen struct {
	module        *ir.Module
	sc            scope
	rt            runtimeFuncs
	funcMap       map[string]*ir.Func
	mainFunc      *ir.Func
	currentFunc   *ir.Func
	currentBlock  *ir.Block
	breakStack    []*ir.Block
	continueStack []*ir.Block
	errors        []error
	blockCounter  int

	funcDecls   map[string]*ast.FnDef
	funcTypes   map[string]vtypes.Type
	structDecls map[string]*ast.StructDef
}

func New() *Codegen {
	return &Codegen{
		module:      ir.NewModule(),
		funcMap:     make(map[string]*ir.Func),
		funcDecls:   make(map[string]*ast.FnDef),
		funcTypes:   make(map[string]vtypes.Type),
		structDecls: make(map[string]*ast.StructDef),
	}
}

func (c *Codegen) Errors() []error { return c.errors }

func (c *Codegen) addError(pos common.Position, format string, args ...interface{}) {
	c.errors = append(c.errors, common.NewError("codegen", pos, fmt.Sprintf(format, args...)))
}

func (c *Codegen) freshBlock(name string) *ir.Block {
	c.blockCounter++
	return ir.NewBlock(fmt.Sprintf("%s_%d", name, c.blockCounter))
}

func (c *Codegen) freshLocal(name string) ir.LocalIdent {
	c.blockCounter++
	return ir.LocalIdent{LocalName: fmt.Sprintf("%s_%d", name, c.blockCounter)}
}

func (c *Codegen) setBlock(block *ir.Block) {
	c.currentBlock = block
	c.currentFunc.Blocks = append(c.currentFunc.Blocks, block)
}

func (c *Codegen) Compile(prog *ast.Program) *ir.Module {
	c.module = ir.NewModule()
	c.collectDecls(prog)
	c.declareRuntimeFuncs()
	c.emitFuncDecls()

	hasMain := c.funcDecls["main"] != nil

	if !hasMain {
		c.mainFunc = c.module.NewFunc("main", lltypes.I32)
		c.funcMap["main"] = c.mainFunc
		c.currentFunc = c.mainFunc
		entry := c.freshBlock("entry")
		c.setBlock(entry)
	}

	c.emitFuncDefs()

	if !hasMain {
		c.currentFunc = c.mainFunc
		entry := c.mainFunc.Blocks[0]
		c.currentBlock = entry
		c.sc.push()
		for _, stmt := range prog.Stmts {
			switch stmt.(type) {
			case *ast.FnDef, *ast.StructDef:
				continue
			}
			if c.currentBlock != nil && c.currentBlock.Term != nil {
				break
			}
			c.codegenStmt(stmt)
		}
		if c.currentBlock != nil && c.currentBlock.Term == nil {
			c.emitTermRet(constant.NewInt(lltypes.I32, 0))
		}
	}

	return c.module
}

func (c *Codegen) collectDecls(prog *ast.Program) {
	for _, stmt := range prog.Stmts {
		switch s := stmt.(type) {
		case *ast.FnDef:
			c.funcDecls[s.Name.String()] = s
		case *ast.StructDef:
			c.structDecls[s.Name.String()] = s
		}
	}
}

func (c *Codegen) emitFuncDecls() {
	for name, fd := range c.funcDecls {
		paramTypes := make([]lltypes.Type, len(fd.Params))
		for i, p := range fd.Params {
			pt := c.resolveTypeExpr(p.Type)
			paramTypes[i] = toLLVMType(pt)
		}
		var retType lltypes.Type = lltypes.Void
		if fd.Return != nil {
			rt := c.resolveTypeExpr(fd.Return)
			retType = toLLVMType(rt)
		}
		params := make([]*ir.Param, len(fd.Params))
		for i, p := range fd.Params {
			params[i] = ir.NewParam(p.Name.String(), paramTypes[i])
		}
		fn := c.module.NewFunc(name, retType, params...)
		c.funcMap[name] = fn

		pts := make([]vtypes.Type, len(fd.Params))
		for i, p := range fd.Params {
			pts[i] = c.resolveTypeExpr(p.Type)
		}
		var retTy vtypes.Type = vtypes.TypeVoid
		if fd.Return != nil {
			retTy = c.resolveTypeExpr(fd.Return)
		}
		c.funcTypes[name] = &vtypes.FnType{ParamTypes: pts, ReturnType: retTy}
	}
}

func (c *Codegen) emitFuncDefs() {
	for name, fd := range c.funcDecls {
		fn := c.funcMap[name]
		c.currentFunc = fn
		entry := c.freshBlock("entry")
		c.setBlock(entry)
		c.sc.push()

		for i, p := range fd.Params {
			pt := c.resolveTypeExpr(p.Type)
			llvmTyp := toLLVMType(pt)
			alloca := ir.NewAlloca(llvmTyp)
			alloca.LocalIdent = c.freshLocal(p.Name.String())
			c.currentBlock.Insts = append(c.currentBlock.Insts, alloca)
			store := ir.NewStore(fn.Params[i], alloca)
			c.currentBlock.Insts = append(c.currentBlock.Insts, store)
			c.sc.set(p.Name.String(), &varInfo{alloca: alloca, typ: pt})
		}

		for _, stmt := range fd.Body.Stmts {
			if c.currentBlock != nil && c.currentBlock.Term != nil {
				break
			}
			c.codegenStmt(stmt)
		}

		var retType vtypes.Type = vtypes.TypeVoid
		if fd.Return != nil {
			retType = c.resolveTypeExpr(fd.Return)
		}
		if c.currentBlock != nil && c.currentBlock.Term == nil {
			if retType.Kind() == common.TypeVoid {
				c.emitTermRet(nil)
			} else {
				llvmRet := toLLVMType(retType)
				c.emitTermRet(constant.NewZeroInitializer(llvmRet))
			}
		}

		c.sc.pop()
	}
}

func (c *Codegen) resolveTypeExpr(expr ast.Expr) vtypes.Type {
	switch e := expr.(type) {
	case *ast.TypeExpr:
		return resolvePrimitiveType(e.Name)
	case *ast.TensorTypeExpr:
		elemType := c.resolveTypeExpr(e.ElemType)
		var shape []int64
		for _, d := range e.Dims {
			switch dim := d.(type) {
			case *ast.NumberLit:
				var val int64
				fmt.Sscanf(dim.Value, "%d", &val)
				shape = append(shape, val)
			case *ast.ArrayLit:
				for _, elem := range dim.Elems {
					if num, ok := elem.(*ast.NumberLit); ok {
						var val int64
						fmt.Sscanf(num.Value, "%d", &val)
						shape = append(shape, val)
					} else {
						shape = append(shape, -1)
					}
				}
			default:
				shape = append(shape, -1)
			}
		}
		return &vtypes.TensorType{ElemType: elemType, Shape: shape}
	case *ast.Ident:
		if vi, ok := c.sc.get(e.Name); ok {
			return vi.typ
		}
		if ft, ok := c.funcTypes[e.Name]; ok {
			return ft
		}
		if sd, ok := c.structDecls[e.Name]; ok {
			fields := make(map[string]vtypes.Type)
			for _, f := range sd.Fields {
				fields[f.Name.String()] = c.resolveTypeExpr(f.Type)
			}
			return &vtypes.StructType{Name: sd.Name.String(), Fields: fields}
		}
	default:
		_ = e
	}
	return vtypes.TypeVoid
}

func resolvePrimitiveType(name string) vtypes.Type {
	switch name {
	case "i8":
		return vtypes.TypeI8
	case "i16":
		return vtypes.TypeI16
	case "i32":
		return vtypes.TypeI32
	case "i64":
		return vtypes.TypeI64
	case "u8":
		return vtypes.TypeU8
	case "u16":
		return vtypes.TypeU16
	case "u32":
		return vtypes.TypeU32
	case "u64":
		return vtypes.TypeU64
	case "f32":
		return vtypes.TypeF32
	case "f64":
		return vtypes.TypeF64
	case "bool":
		return vtypes.TypeBool
	case "string":
		return vtypes.TypeString
	case "void":
		return vtypes.TypeVoid
	default:
		return vtypes.TypeVoid
	}
}

func (c *Codegen) emitTermRet(val value.Value) {
	c.currentBlock.Term = ir.NewRet(val)
}

func (c *Codegen) emitTermBr(target *ir.Block) {
	c.currentBlock.Term = ir.NewBr(target)
}

func (c *Codegen) emitTermCondBr(cond value.Value, thenBlock, elseBlock *ir.Block) {
	c.currentBlock.Term = ir.NewCondBr(cond, thenBlock, elseBlock)
}

func hasDecimal(s string) bool {
	return strings.Contains(s, ".") || strings.Contains(s, "e") || strings.Contains(s, "E")
}
