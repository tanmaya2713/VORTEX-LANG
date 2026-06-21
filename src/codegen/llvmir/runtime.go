package llvmir

import (
	"github.com/llir/llvm/ir"
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/llir/llvm/ir/value"
)

type runtimeFuncs struct {
	printf              value.Value
	vortexInit          value.Value
	vortexTensorMatmul  value.Value
	vortexTensorAdd     value.Value
	vortexTensorRelu    value.Value
	vortexTensorSigmoid value.Value
	vortexTensorCreate  value.Value
	vortexPrintI32      value.Value
	vortexPrintF64      value.Value
	vortexPrintBool     value.Value
	vortexPrintStr      value.Value
	vortexPrintNewline  value.Value
}

func (c *Codegen) declareRuntimeFuncs() {
	m := c.module
	printf := m.NewFunc("printf", lltypes.I32,
		ir.NewParam("fmt", lltypes.I8Ptr),
	)
	printf.Sig.Variadic = true
	c.rt.printf = printf

	c.rt.vortexInit = m.NewFunc("vortex_init", lltypes.Void)

	tensorPtr := lltypes.NewPointer(getTensorStructType())

	c.rt.vortexTensorCreate = m.NewFunc("vortex_tensor_create", tensorPtr,
		ir.NewParam("shape", lltypes.NewPointer(lltypes.I32)),
		ir.NewParam("ndim", lltypes.I32),
		ir.NewParam("data", lltypes.NewPointer(lltypes.Float)),
	)

	c.rt.vortexTensorMatmul = m.NewFunc("vortex_matmul", lltypes.Void,
		ir.NewParam("a", tensorPtr), ir.NewParam("b", tensorPtr),
		ir.NewParam("out", tensorPtr),
	)

	c.rt.vortexTensorAdd = m.NewFunc("vortex_tensor_add", lltypes.Void,
		ir.NewParam("a", tensorPtr), ir.NewParam("b", tensorPtr),
		ir.NewParam("out", tensorPtr),
	)

	c.rt.vortexTensorRelu = m.NewFunc("vortex_tensor_relu", lltypes.Void,
		ir.NewParam("a", tensorPtr),
		ir.NewParam("out", tensorPtr),
	)

	c.rt.vortexTensorSigmoid = m.NewFunc("vortex_tensor_sigmoid", lltypes.Void,
		ir.NewParam("a", tensorPtr),
		ir.NewParam("out", tensorPtr),
	)

	c.rt.vortexPrintI32 = m.NewFunc("vortex_print_i32", lltypes.Void,
		ir.NewParam("val", lltypes.I32),
	)

	c.rt.vortexPrintF64 = m.NewFunc("vortex_print_f64", lltypes.Void,
		ir.NewParam("val", lltypes.Double),
	)

	c.rt.vortexPrintBool = m.NewFunc("vortex_print_bool", lltypes.Void,
		ir.NewParam("val", lltypes.I1),
	)

	c.rt.vortexPrintStr = m.NewFunc("vortex_print_string", lltypes.Void,
		ir.NewParam("str", lltypes.I8Ptr),
	)

	c.rt.vortexPrintNewline = m.NewFunc("vortex_print_newline", lltypes.Void)
}
