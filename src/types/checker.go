package types

import (
	"fmt"
	"strings"

	"github.com/vortex-lang/vortex/src/ast"
	"github.com/vortex-lang/vortex/src/common"
	"github.com/vortex-lang/vortex/src/dict"
)

type Scope struct {
	vars   map[string]Type
	parent *Scope
}

func newScope(parent *Scope) *Scope {
	return &Scope{vars: make(map[string]Type), parent: parent}
}

func (s *Scope) lookup(name string) (Type, bool) {
	t, ok := s.vars[name]
	if !ok && s.parent != nil {
		return s.parent.lookup(name)
	}
	return t, ok
}

func (s *Scope) declare(name string, t Type) error {
	if _, exists := s.vars[name]; exists {
		return fmt.Errorf("variable '%s' already declared in this scope", name)
	}
	s.vars[name] = t
	return nil
}

type Checker struct {
	globals     *Scope
	funcs       map[string]*FnType
	funcDecls   map[string]*ast.FnDef
	structDecls map[string]*ast.StructDef
	errs        []error
}

func New() *Checker {
	return &Checker{
		globals:     newScope(nil),
		funcs:       make(map[string]*FnType),
		funcDecls:   make(map[string]*ast.FnDef),
		structDecls: make(map[string]*ast.StructDef),
	}
}

func (c *Checker) Errors() []error { return c.errs }

func (c *Checker) addError(pos common.Position, format string, args ...interface{}) {
	c.errs = append(c.errs, common.NewError("typecheck", pos, fmt.Sprintf(format, args...)))
}

func (c *Checker) Check(prog *ast.Program) bool {
	c.collectDecls(prog)
	c.checkProgram(prog)
	return len(c.errs) == 0
}

func (c *Checker) collectDecls(prog *ast.Program) {
	for _, stmt := range prog.Stmts {
		switch s := stmt.(type) {
		case *ast.FnDef:
			fnType := c.typeFromFnDef(s)
			c.funcs[s.Name.String()] = fnType
			c.funcDecls[s.Name.String()] = s
		case *ast.StructDef:
			c.structDecls[s.Name.String()] = s
		}
	}
}

func (c *Checker) typeFromFnDef(fn *ast.FnDef) *FnType {
	var paramTypes []Type
	for _, p := range fn.Params {
		t := c.typeFromExprType(p.Type)
		paramTypes = append(paramTypes, t)
	}
	var retType Type = TypeVoid
	if fn.Return != nil {
		retType = c.typeFromExprType(fn.Return)
	}
	return &FnType{ParamTypes: paramTypes, ReturnType: retType}
}

func (c *Checker) typeFromExprType(expr ast.Expr) Type {
	if t, ok := expr.(*ast.TypeExpr); ok {
		switch t.Name {
		case "i8":
			return TypeI8
		case "i16":
			return TypeI16
		case "i32":
			return TypeI32
		case "i64":
			return TypeI64
		case "u8":
			return TypeU8
		case "u16":
			return TypeU16
		case "u32":
			return TypeU32
		case "u64":
			return TypeU64
		case "f32":
			return TypeF32
		case "f64":
			return TypeF64
		case "bool":
			return TypeBool
		case "string":
			return TypeString
		case "void":
			return TypeVoid
		}
		if dict.IsOperatorKeyword(t.Name) {
			return &LayerType{LayerKind: t.Name}
		}
		if _, ok := dict.LookupKeyword(t.Name); ok {
			return &LayerType{LayerKind: t.Name}
		}
		if structType, ok := c.structDecls[t.Name]; ok {
			fields := make(map[string]Type)
			for _, f := range structType.Fields {
				fields[f.Name.String()] = c.typeFromExprType(f.Type)
			}
			return &StructType{Name: t.Name, Fields: fields}
		}
		return &NamedType{Name: t.Name}
	}
	if tensor, ok := expr.(*ast.TensorTypeExpr); ok {
		elemType := c.typeFromExprType(tensor.ElemType)
		var shape []int64
		for _, dim := range tensor.Dims {
			switch d := dim.(type) {
			case *ast.NumberLit:
				var val int64
				fmt.Sscanf(d.Value, "%d", &val)
				shape = append(shape, val)
			case *ast.ArrayLit:
				for _, elem := range d.Elems {
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
		return &TensorType{ElemType: elemType, Shape: shape}
	}
	return TypeError
}

func (c *Checker) checkProgram(prog *ast.Program) {
	for _, stmt := range prog.Stmts {
		c.checkStmt(stmt, c.globals)
	}
}

func (c *Checker) checkStmt(stmt ast.Stmt, scope *Scope) Type {
	switch s := stmt.(type) {
	case *ast.LetStmt:
		return c.checkLet(s, scope)
	case *ast.FnDef:
		return c.checkFnDef(s)
	case *ast.StructDef:
		return c.checkStructDef(s)
	case *ast.IfStmt:
		return c.checkIf(s, scope)
	case *ast.ForStmt:
		return c.checkFor(s, scope)
	case *ast.WhileStmt:
		return c.checkWhile(s, scope)
	case *ast.ReturnStmt:
		return c.checkReturn(s, scope)
	case *ast.BreakStmt:
		return TypeVoid
	case *ast.ContinueStmt:
		return TypeVoid
	case *ast.BlockStmt:
		return c.checkBlock(s, scope)
	case *ast.ExprStmt:
		return c.checkExpr(s.E, scope)
	case *ast.PrintStmt:
		c.checkExpr(s.Expr, scope)
		return TypeVoid
	case *ast.AssertStmt:
		condType := c.checkExpr(s.Cond, scope)
		if !isBool(condType) {
			c.addError(stmt.Pos(), "assert condition must be bool, got %s", condType)
		}
		return TypeVoid
	case *ast.ImportStmt:
		return TypeVoid
	case *ast.ModelDef:
		return c.checkModelDef(s)
	case *ast.TrainStmt:
		return c.checkTrainStmt(s, scope)
	case *ast.AssignExpr:
		return c.checkAssign(s, scope)
	default:
		return TypeError
	}
}

func (c *Checker) checkLet(s *ast.LetStmt, scope *Scope) Type {
	valType := c.checkExpr(s.Value, scope)
	if s.Type != nil {
		annType := c.typeFromExprType(s.Type)
		if !c.assignableTo(valType, annType) {
			c.addError(s.Pos(), "cannot assign %s to %s", valType, annType)
		}
		valType = annType
	}
	if err := scope.declare(s.Name.String(), valType); err != nil {
		c.addError(s.Pos(), "%s", err.Error())
	}
	return valType
}

func (c *Checker) checkStructDef(s *ast.StructDef) Type {
	fields := make(map[string]Type)
	for _, f := range s.Fields {
		ft := c.typeFromExprType(f.Type)
		fields[f.Name.String()] = ft
	}
	if _, exists := c.structDecls[s.Name.String()]; !exists {
		c.structDecls[s.Name.String()] = s
	}
	return &StructType{Name: s.Name.String(), Fields: fields}
}

func (c *Checker) checkFnDef(s *ast.FnDef) Type {
	fnType := c.funcs[s.Name.String()]
	if fnType == nil {
		c.addError(s.Pos(), "function '%s' not collected", s.Name)
		return TypeError
	}
	fnScope := newScope(c.globals)
	for i, p := range s.Params {
		pt := fnType.ParamTypes[i]
		fnScope.declare(p.Name.String(), pt)
	}
	var retType Type
	for _, stmt := range s.Body.Stmts {
		retType = c.checkStmt(stmt, fnScope)
	}
	if len(s.Body.Stmts) == 0 {
		retType = TypeVoid
	}
	if s.Return != nil {
		declaredRet := c.typeFromExprType(s.Return)
		if retType != TypeVoid && !c.assignableTo(retType, declaredRet) {
			c.addError(s.Pos(), "return type mismatch: expected %s, got %s", declaredRet, retType)
		}
	}
	return fnType
}

func (c *Checker) checkIf(s *ast.IfStmt, scope *Scope) Type {
	condType := c.checkExpr(s.Cond, scope)
	if !isBool(condType) {
		c.addError(s.Cond.Pos(), "if condition must be bool, got %s", condType)
	}
	thenType := c.checkBlock(s.Then, scope)
	if s.Else != nil {
		elseType := c.checkStmt(s.Else, scope)
		if isAssignable(thenType, elseType) {
			return elseType
		}
	}
	return thenType
}

func (c *Checker) checkFor(s *ast.ForStmt, scope *Scope) Type {
	rangeType := c.checkExpr(s.Range, scope)
	if rangeType.Kind() != common.TypeArray && rangeType.Kind() != common.TypeTensor {
		c.addError(s.Range.Pos(), "for range must be array or tensor, got %s", rangeType)
	}
	loopScope := newScope(scope)
	var elemType Type = TypeI32
	if at, ok := rangeType.(*ArrayType); ok {
		elemType = at.ElemType
	} else if tt, ok := rangeType.(*TensorType); ok {
		elemType = tt.ElemType
	}
	loopScope.declare(s.Var.String(), elemType)
	return c.checkBlock(s.Body, loopScope)
}

func (c *Checker) checkWhile(s *ast.WhileStmt, scope *Scope) Type {
	condType := c.checkExpr(s.Cond, scope)
	if !isBool(condType) {
		c.addError(s.Cond.Pos(), "while condition must be bool, got %s", condType)
	}
	return c.checkBlock(s.Body, newScope(scope))
}

func (c *Checker) checkReturn(s *ast.ReturnStmt, scope *Scope) Type {
	if s.Value != nil {
		return c.checkExpr(s.Value, scope)
	}
	return TypeVoid
}

func (c *Checker) checkBlock(s *ast.BlockStmt, scope *Scope) Type {
	blockScope := newScope(scope)
	var lastType Type = TypeVoid
	for _, stmt := range s.Stmts {
		lastType = c.checkStmt(stmt, blockScope)
	}
	return lastType
}

func (c *Checker) checkModelDef(s *ast.ModelDef) Type {
	for _, layer := range s.Layers {
		if id, ok := layer.Kind.(*ast.Ident); ok {
			if !dict.IsOperatorKeyword(id.String()) {
				c.addError(layer.Pos(), "unknown layer type '%s'", id)
			}
		}
		for _, param := range layer.Params {
			if id, ok := param.Value.(*ast.Ident); ok && dict.IsKeyword(id.String()) {
				continue
			}
			c.checkExpr(param.Value, c.globals)
		}
	}
	return &ModelType{Name: s.Name.String()}
}

func (c *Checker) checkTrainStmt(s *ast.TrainStmt, scope *Scope) Type {
	c.checkExpr(s.Model, scope)
	c.checkExpr(s.Data, scope)
	if s.Epochs != nil {
		epochType := c.checkExpr(s.Epochs, scope)
		if !isInteger(epochType) {
			c.addError(s.Epochs.Pos(), "epochs must be integer, got %s", epochType)
		}
	}
	if s.LR != nil {
		lrType := c.checkExpr(s.LR, scope)
		if !isFloat(lrType) && !isInteger(lrType) {
			c.addError(s.LR.Pos(), "learning rate must be numeric, got %s", lrType)
		}
	}
	return TypeVoid
}

func (c *Checker) checkExpr(expr ast.Expr, scope *Scope) Type {
	switch e := expr.(type) {
	case *ast.Ident:
		return c.checkIdent(e, scope)
	case *ast.NumberLit:
		return c.checkNumber(e)
	case *ast.StringLit:
		return TypeString
	case *ast.BoolLit:
		return TypeBool
	case *ast.BinaryExpr:
		return c.checkBinary(e, scope)
	case *ast.UnaryExpr:
		return c.checkUnary(e, scope)
	case *ast.CallExpr:
		return c.checkCall(e, scope)
	case *ast.IndexExpr:
		return c.checkIndex(e, scope)
	case *ast.MemberExpr:
		return c.checkMember(e, scope)
	case *ast.AssignExpr:
		return c.checkAssign(e, scope)
	case *ast.TypeExpr:
		return c.typeFromExprType(e)
	case *ast.TensorTypeExpr:
		return c.typeFromExprType(e)
	case *ast.ArrayLit:
		return c.checkArrayLit(e, scope)
	default:
		return TypeError
	}
}

func (c *Checker) checkIdent(e *ast.Ident, scope *Scope) Type {
	t, ok := scope.lookup(e.String())
	if !ok {
		if ft, exists := c.funcs[e.String()]; exists {
			return ft
		}
		if st, exists := c.structDecls[e.String()]; exists {
			fields := make(map[string]Type)
			for _, f := range st.Fields {
				fields[f.Name.String()] = c.typeFromExprType(f.Type)
			}
			return &StructType{Name: e.String(), Fields: fields}
		}
		c.addError(e.Pos(), "undefined variable '%s'", e)
		return TypeError
	}
	return t
}

func (c *Checker) checkNumber(e *ast.NumberLit) Type {
	if e.Kind == "" {
		if strings.Contains(e.Value, ".") {
			return TypeF64
		}
		return TypeI32
	}
	switch e.Kind {
	case "i8":
		return TypeI8
	case "i16":
		return TypeI16
	case "i32":
		return TypeI32
	case "i64":
		return TypeI64
	case "u8":
		return TypeU8
	case "u16":
		return TypeU16
	case "u32":
		return TypeU32
	case "u64":
		return TypeU64
	case "f32":
		return TypeF32
	case "f64":
		return TypeF64
	}
	return TypeI32
}

func (c *Checker) checkBinary(e *ast.BinaryExpr, scope *Scope) Type {
	leftType := c.checkExpr(e.Left, scope)
	rightType := c.checkExpr(e.Right, scope)

	switch e.Op {
	case "+", "-", "*", "/", "%":
		if isNumeric(leftType) && isNumeric(rightType) {
			return commonType(leftType, rightType)
		}
		if e.Op == "+" && isString(leftType) && isString(rightType) {
			return TypeString
		}
		if e.Op == "+" && leftType.Kind() == common.TypeTensor && rightType.Kind() == common.TypeTensor {
			ltt, lok := leftType.(*TensorType)
			rtt, rok := rightType.(*TensorType)
			if lok && rok && c.shapesEqual(ltt.Shape, rtt.Shape) {
				return &TensorType{ElemType: ltt.ElemType, Shape: ltt.Shape}
			}
			c.addError(e.Pos(), "tensor add requires same shape, got %s and %s", leftType, rightType)
			return TypeError
		}
		if e.Op == "*" && leftType.Kind() == common.TypeTensor && rightType.Kind() == common.TypeTensor {
			if ltt, ok := leftType.(*TensorType); ok && len(ltt.Shape) >= 2 {
				if rtt, ok2 := rightType.(*TensorType); ok2 && len(rtt.Shape) >= 2 {
					outShape := make([]int64, 2)
					outShape[0] = ltt.Shape[0]
					outShape[1] = rtt.Shape[1]
					return &TensorType{ElemType: ltt.ElemType, Shape: outShape}
				}
			}
			return leftType
		}
		c.addError(e.Pos(), "operator '%s' requires numeric operands, got %s and %s", e.Op, leftType, rightType)
	case "==", "!=", "<", "<=", ">", ">=":
		if isNumeric(leftType) && isNumeric(rightType) {
			return TypeBool
		}
		if isString(leftType) && isString(rightType) && (e.Op == "==" || e.Op == "!=") {
			return TypeBool
		}
		c.addError(e.Pos(), "comparison '%s' requires compatible operands, got %s and %s", e.Op, leftType, rightType)
	case "&&", "||":
		if !isBool(leftType) || !isBool(rightType) {
			c.addError(e.Pos(), "operator '%s' requires bool operands, got %s and %s", e.Op, leftType, rightType)
		}
		return TypeBool
	default:
		c.addError(e.Pos(), "unknown operator '%s'", e.Op)
	}
	return TypeError
}

func (c *Checker) checkUnary(e *ast.UnaryExpr, scope *Scope) Type {
	rightType := c.checkExpr(e.Right, scope)
	switch e.Op {
	case "-":
		if !isNumeric(rightType) {
			c.addError(e.Pos(), "negation requires numeric operand, got %s", rightType)
		}
		return rightType
	case "!":
		if !isBool(rightType) {
			c.addError(e.Pos(), "logical not requires bool operand, got %s", rightType)
		}
		return TypeBool
	}
	return TypeError
}

func (c *Checker) shapesEqual(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (c *Checker) checkCall(e *ast.CallExpr, scope *Scope) Type {
	if ident, ok := e.Fn.(*ast.Ident); ok {
		switch ident.Name {
		case "relu", "sigmoid":
			if len(e.Args) != 1 {
				c.addError(e.Pos(), "%s requires exactly 1 argument", ident.Name)
				return TypeError
			}
			argType := c.checkExpr(e.Args[0], scope)
			if argType.Kind() != common.TypeTensor {
				c.addError(e.Pos(), "%s requires a tensor argument, got %s", ident.Name, argType)
				return TypeError
			}
			return argType
		}
	}
	fnType := c.checkExpr(e.Fn, scope)
	if fnType.Kind() != common.TypeFn {
		c.addError(e.Pos(), "cannot call non-function type %s", fnType)
		return TypeError
	}
	ft, ok := fnType.(*FnType)
	if !ok {
		c.addError(e.Pos(), "expected function type")
		return TypeError
	}
	if len(e.Args) != len(ft.ParamTypes) {
		c.addError(e.Pos(), "expected %d arguments, got %d", len(ft.ParamTypes), len(e.Args))
		return ft.ReturnType
	}
	for i, arg := range e.Args {
		argType := c.checkExpr(arg, scope)
		if !c.assignableTo(argType, ft.ParamTypes[i]) {
			c.addError(arg.Pos(), "argument %d: expected %s, got %s", i+1, ft.ParamTypes[i], argType)
		}
	}
	return ft.ReturnType
}

func (c *Checker) checkIndex(e *ast.IndexExpr, scope *Scope) Type {
	targetType := c.checkExpr(e.Target, scope)
	indexType := c.checkExpr(e.Index, scope)

	if !isInteger(indexType) {
		c.addError(e.Index.Pos(), "index must be integer, got %s", indexType)
		return TypeError
	}

	switch t := targetType.(type) {
	case *ArrayType:
		return t.ElemType
	case *TensorType:
		if len(t.Shape) > 0 {
			dims := t.Shape
			if len(dims) == 1 {
				return t.ElemType
			}
			newShape := dims[1:]
			return &TensorType{ElemType: t.ElemType, Shape: newShape}
		}
		return t.ElemType
	default:
		c.addError(e.Target.Pos(), "cannot index %s", targetType)
		return TypeError
	}
}

func (c *Checker) checkMember(e *ast.MemberExpr, scope *Scope) Type {
	targetType := c.checkExpr(e.Target, scope)
	switch t := targetType.(type) {
	case *StructType:
		fieldType, ok := t.Fields[e.Name.String()]
		if !ok {
			c.addError(e.Pos(), "struct '%s' has no field '%s'", t.Name, e.Name)
			return TypeError
		}
		return fieldType
	case *ModelType:
		return targetType
	default:
		c.addError(e.Target.Pos(), "type %s has no fields", targetType)
		return TypeError
	}
}

func (c *Checker) checkAssign(e *ast.AssignExpr, scope *Scope) Type {
	leftType := c.checkExpr(e.Left, scope)
	rightType := c.checkExpr(e.Right, scope)
	if !c.assignableTo(rightType, leftType) {
		c.addError(e.Pos(), "cannot assign %s to variable of type %s", rightType, leftType)
	}
	return leftType
}

func (c *Checker) checkArrayLit(e *ast.ArrayLit, scope *Scope) Type {
	if len(e.Elems) == 0 {
		return &ArrayType{ElemType: TypeError, Len: 0}
	}
	elemType := c.checkExpr(e.Elems[0], scope)
	for _, elem := range e.Elems[1:] {
		et := c.checkExpr(elem, scope)
		if !isAssignable(elemType, et) {
			elemType = commonType(elemType, et)
		}
	}
	return &ArrayType{ElemType: elemType, Len: int64(len(e.Elems))}
}

func (c *Checker) assignableTo(from, to Type) bool {
	if from.Kind() == common.TypeError || to.Kind() == common.TypeError {
		return true
	}
	return isAssignable(from, to)
}

func isAssignable(a, b Type) bool {
	if a == TypeError || b == TypeError {
		return true
	}
	if a.Kind() == b.Kind() {
		if a.Kind() == common.TypeArray {
			at, aok := a.(*ArrayType)
			bt, bok := b.(*ArrayType)
			if aok && bok {
				return at.Len == 0 || bt.Len == 0 || at.Len == bt.Len
			}
		}
		return true
	}
	if isNumericKind(a.Kind()) && isNumericKind(b.Kind()) {
		return true
	}
	if b.Kind() == common.TypeTensor && a.Kind() == common.TypeArray {
		at, aok := a.(*ArrayType)
		bt, bok := b.(*TensorType)
		if aok && bok {
			total := int64(1)
			for _, d := range bt.Shape {
				total *= d
			}
			return at.Len == total
		}
		return true
	}
	return false
}

func isNumericKind(k common.DataType) bool {
	switch k {
	case common.TypeI8, common.TypeI16, common.TypeI32, common.TypeI64,
		common.TypeU8, common.TypeU16, common.TypeU32, common.TypeU64,
		common.TypeF32, common.TypeF64:
		return true
	}
	return false
}

func commonType(a, b Type) Type {
	if a.Kind() == common.TypeError || b.Kind() == common.TypeError {
		return TypeError
	}
	ranks := map[common.DataType]int{
		common.TypeI8:  0,
		common.TypeI16: 1,
		common.TypeI32: 2,
		common.TypeI64: 3,
		common.TypeU8:  4,
		common.TypeU16: 5,
		common.TypeU32: 6,
		common.TypeU64: 7,
		common.TypeF32: 8,
		common.TypeF64: 9,
	}
	ra, aok := ranks[a.Kind()]
	rb, bok := ranks[b.Kind()]
	if !aok || !bok {
		return a
	}
	if ra >= rb {
		return a
	}
	return b
}

func isNumeric(t Type) bool {
	switch t.Kind() {
	case common.TypeI8, common.TypeI16, common.TypeI32, common.TypeI64,
		common.TypeU8, common.TypeU16, common.TypeU32, common.TypeU64,
		common.TypeF32, common.TypeF64:
		return true
	}
	return false
}

func isInteger(t Type) bool {
	switch t.Kind() {
	case common.TypeI8, common.TypeI16, common.TypeI32, common.TypeI64,
		common.TypeU8, common.TypeU16, common.TypeU32, common.TypeU64:
		return true
	}
	return false
}

func isFloat(t Type) bool {
	return t.Kind() == common.TypeF32 || t.Kind() == common.TypeF64
}

func isBool(t Type) bool {
	return t.Kind() == common.TypeBool
}

func isString(t Type) bool {
	return t.Kind() == common.TypeString
}
