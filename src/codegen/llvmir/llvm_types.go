package llvmir

import (
	lltypes "github.com/llir/llvm/ir/types"
	"github.com/vortex-lang/vortex/src/common"
	vtypes "github.com/vortex-lang/vortex/src/types"
)

var tensorStructType = func() lltypes.Type {
	st := lltypes.NewStruct(
		lltypes.NewPointer(lltypes.I32),
		lltypes.I32,
		lltypes.NewPointer(lltypes.Float),
	)
	st.SetName("VortexTensor")
	return st
}()

func toLLVMType(t vtypes.Type) lltypes.Type {
	switch t.Kind() {
	case common.TypeVoid:
		return lltypes.Void
	case common.TypeI8, common.TypeU8:
		return lltypes.I8
	case common.TypeI16, common.TypeU16:
		return lltypes.I16
	case common.TypeI32, common.TypeU32:
		return lltypes.I32
	case common.TypeI64, common.TypeU64:
		return lltypes.I64
	case common.TypeF32:
		return lltypes.Float
	case common.TypeF64:
		return lltypes.Double
	case common.TypeBool:
		return lltypes.I1
	case common.TypeString:
		return lltypes.I8Ptr
	case common.TypeTensor:
		return lltypes.NewPointer(tensorStructType)
	case common.TypeModel, common.TypeLayer, common.TypeStruct:
		return lltypes.I8Ptr
	case common.TypeArray:
		if at, ok := t.(*vtypes.ArrayType); ok && at.Len > 0 {
			elemLLVM := toLLVMType(at.ElemType)
			return lltypes.NewArray(uint64(at.Len), elemLLVM)
		}
		return lltypes.I8Ptr
	case common.TypeFn:
		return lltypes.I8Ptr
	case common.TypePtr:
		if pt, ok := t.(*vtypes.PtrType); ok {
			return lltypes.NewPointer(toLLVMType(pt.ElemType))
		}
		return lltypes.I8Ptr
	default:
		return lltypes.I8Ptr
	}
}
