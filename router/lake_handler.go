package router

type LakeHandler struct {
}

func ProvideLakeHandler() *LakeHandler {
	return &LakeHandler{}
}
