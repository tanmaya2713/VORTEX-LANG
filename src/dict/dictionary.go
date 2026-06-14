package dict

type Keyword int

const (
	KwLet Keyword = iota
	KwFn
	KwIf
	KwElse
	KwFor
	KwWhile
	KwReturn
	KwBreak
	KwContinue
	KwMut
	KwRef
	KwStruct
	KwModel
	KwLayer
	KwTrain
	KwEpochs
	KwLR
	KwStrategy
	KwDevices
	KwDense
	KwConv2D
	KwConv3D
	KwMaxPool
	KwAvgPool
	KwDropout
	KwBatchNorm
	KwLayerNorm
	KwRelu
	KwSigmoid
	KwTanh
	KwSoftmax
	KwGelu
	KwCrossEntropy
	KwMSE
	KwMAE
	KwHuber
	KwAdam
	KwSGD
	KwRMSProp
	KwAdamW
	KwTrue
	KwFalse
	KwPrint
	KwAssert
	KwImport
	KwType
	KwTensor
)

var KeywordNames = map[Keyword]string{
	KwLet:          "let",
	KwFn:           "fn",
	KwIf:           "if",
	KwElse:         "else",
	KwFor:          "for",
	KwWhile:        "while",
	KwReturn:       "return",
	KwBreak:        "break",
	KwContinue:     "continue",
	KwMut:          "mut",
	KwRef:          "ref",
	KwStruct:       "struct",
	KwModel:        "model",
	KwLayer:        "layer",
	KwTrain:        "train",
	KwEpochs:       "epochs",
	KwLR:           "lr",
	KwStrategy:     "strategy",
	KwDevices:      "devices",
	KwDense:        "dense",
	KwConv2D:       "conv2d",
	KwConv3D:       "conv3d",
	KwMaxPool:      "maxpool",
	KwAvgPool:      "avgpool",
	KwDropout:      "dropout",
	KwBatchNorm:    "batchnorm",
	KwLayerNorm:    "layernorm",
	KwRelu:         "relu",
	KwSigmoid:      "sigmoid",
	KwTanh:         "tanh",
	KwSoftmax:      "softmax",
	KwGelu:         "gelu",
	KwCrossEntropy: "cross_entropy",
	KwMSE:          "mse",
	KwMAE:          "mae",
	KwHuber:        "huber",
	KwAdam:         "adam",
	KwSGD:          "sgd",
	KwRMSProp:      "rmsprop",
	KwAdamW:        "adamw",
	KwTrue:         "true",
	KwFalse:        "false",
	KwPrint:        "print",
	KwAssert:       "assert",
	KwImport:       "import",
	KwType:         "type",
	KwTensor:       "tensor",
}

var KeywordMap map[string]Keyword

func init() {
	KeywordMap = make(map[string]Keyword, len(KeywordNames))
	for k, v := range KeywordNames {
		KeywordMap[v] = k
	}
}

func IsKeyword(s string) bool {
	_, ok := KeywordMap[s]
	return ok
}

func LookupKeyword(s string) (Keyword, bool) {
	k, ok := KeywordMap[s]
	return k, ok
}

var NoiseWords = map[string]bool{
	"an": true, "the": true,
	"is": true, "are": true, "was": true, "were": true, "be": true,
	"has": true, "have": true, "had": true,
	"do": true, "does": true, "did": true,
	"will": true, "would": true, "shall": true, "should": true,
	"can": true, "could": true, "may": true, "might": true, "must": true,
	"this": true, "that": true, "these": true, "those": true,
	"it": true, "its": true,
	"of": true, "in": true, "on": true, "at": true, "by": true, "to": true,
	"for": true, "with": true, "from": true, "into": true, "through": true,
	"as": true, "than": true, "then": true,
	"and": true, "or": true, "but": true, "not": true, "nor": true,
	"so": true, "yet": true, "if": true,
	"my": true, "your": true, "our": true, "their": true,
	"me": true, "you": true, "we": true, "they": true,
	"using": true, "trained": true, "over": true, "across": true,
	"via": true, "per": true, "initialized": true, "regularized": true,
	"normalized": true, "distributed": true, "synchronized": true,
	"replicated": true, "partitioned": true, "sharded": true,
	"quantized": true, "pruned": true, "between": true,
	"which": true, "what": true, "where": true, "when": true, "why": true,
	"how": true, "each": true, "every": true, "all": true, "some": true,
	"any": true, "no": true, "none": true, "both": true, "either": true,
	"neither": true, "whose": true, "whom": true, "who": true,
	"about": true, "above": true, "after": true, "against": true,
	"along": true, "among": true, "around": true, "before": true,
	"behind": true, "below": true, "beneath": true, "beside": true,
	"beyond": true, "during": true, "except": true, "inside": true,
	"outside": true, "since": true, "throughout": true,
	"toward": true, "under": true, "until": true, "upon": true, "within": true,
	"without": true, "being": true, "having": true, "doing": true,
	"much": true, "more": true, "most": true, "many": true, "few": true,
	"less": true, "least": true, "same": true, "such": true, "only": true,
	"own": true, "very": true, "just": true, "also": true, "too": true,
	"well": true, "here": true, "there": true, "now": true,
}

func IsNoise(word string) bool {
	return NoiseWords[word]
}

var TypeNames = map[string]bool{
	"void": true, "i8": true, "i16": true, "i32": true, "i64": true,
	"u8": true, "u16": true, "u32": true, "u64": true,
	"f32": true, "f64": true, "bool": true, "string": true,
	"tensor": true, "model": true, "layer": true,
}

func IsTypeName(s string) bool {
	return TypeNames[s]
}

func IsOperatorKeyword(s string) bool {
	switch s {
	case "dense", "conv2d", "conv3d", "maxpool", "avgpool",
		"dropout", "batchnorm", "layernorm",
		"relu", "sigmoid", "tanh", "softmax", "gelu",
		"cross_entropy", "mse", "mae", "huber",
		"adam", "sgd", "rmsprop", "adamw":
		return true
	}
	return false
}
