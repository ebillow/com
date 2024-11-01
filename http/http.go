package http

func NewEngine() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.RemoveExtraSlash = true
	r.Use(gin.Recovery(), cors.Default())
	r.NoMethod(func(c *gin.Context) { Fail(c, 405, netpack.ErrorCode_ErrorCode_Faild, "method not allowed") })
	r.NoRoute(func(c *gin.Context) { Fail(c, 404, netpack.ErrorCode_ErrorCode_Faild, "api not found") })

	r.HEAD("/health", func(c *gin.Context) { c.AbortWithStatus(200) })

	return r
}
