package TeaGo

import (
	"errors"
	"github.com/tjbrains/TeaGo/Tea"
	"github.com/tjbrains/TeaGo/actions"
	"github.com/tjbrains/TeaGo/files"
	"github.com/tjbrains/TeaGo/lists"
	"github.com/tjbrains/TeaGo/logs"
	"github.com/tjbrains/TeaGo/processes"
	"github.com/tjbrains/TeaGo/types"
	"github.com/tjbrains/TeaGo/utils/string"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

// 文本mime-type列表
var textMimeMap = map[string]bool{
	"application/atom+xml":                true,
	"application/javascript":              true,
	"application/x-javascript":            true,
	"application/json":                    true,
	"application/rss+xml":                 true,
	"application/x-web-app-manifest+json": true,
	"application/xhtml+xml":               true,
	"application/xml":                     true,
	"image/svg+xml":                       true,
	"text/css":                            true,
	"text/plain":                          true,
	"text/javascript":                     true,
	"text/xml":                            true,
	"text/html":                           true,
	"text/xhtml":                          true,
	"text/sgml":                           true,
}

// 服务启动之前的要执行的函数
var beforeStartFunctions = []func(server *Server){}
var beforeStartOnce = sync.Once{}

// 服务停止之前的要执行的函数
var beforeStopFunctions = []func(server *Server){}
var beforeStopOnce = sync.Once{}

// BeforeStart 在服务启动之前执行一个函数
func BeforeStart(fn func(server *Server)) {
	beforeStartFunctions = append(beforeStartFunctions, fn)
}

// BeforeStop 在服务停止之前执行一个函数
func BeforeStop(fn func(server *Server)) {
	beforeStopFunctions = append(beforeStopFunctions, fn)
}

var actionPathRegexp = regexp.MustCompile(`:(\w+)(\s*(\([^)]+\))?)`)

// Server Web服务
type Server struct {
	singleInstance bool

	directRoutes  map[string]func(writer http.ResponseWriter, request *http.Request)
	patternRoutes []ServerRoutePattern
	routerLocker  sync.Mutex
	staticDirs    []ServerStaticDir

	sessionManager    any
	sessionCookieName string

	lastModule  string       //当前的模块
	lastPrefix  string       //当前的URL前缀
	lastHelpers []any        // 当前的Helper列表
	lastData    actions.Data // 当前的变量列表

	config    *ServerConfig
	logWriter LogWriter
	accessLog bool // 是否记录访问日志

	httpServers      []*http.Server
	httpServerLocker sync.Mutex
	connState        func(conn net.Conn, state http.ConnState)

	internalErrorLogger *log.Logger
	readHeaderTimeout   time.Duration
	readTimeout         time.Duration

	beforeRequestFunc func(writer http.ResponseWriter, req *http.Request) bool
}

// ServerRoutePattern 路由配置
type ServerRoutePattern struct {
	module  string
	reg     *regexp.Regexp
	names   []string
	method  string
	runFunc func(writer http.ResponseWriter, request *http.Request)
}

// ServerStaticDir 静态资源目录
type ServerStaticDir struct {
	prefix string
	dir    string
}

// NewServer 构建一个新的Server
func NewServer(singleInstance ...bool) *Server {
	var server = &Server{
		accessLog: true,
	}

	if len(singleInstance) == 0 {
		server.singleInstance = true
	} else {
		server.singleInstance = singleInstance[0]
	}

	server.init()

	return server
}

// 初始化
func (this *Server) init() {
	this.directRoutes = make(map[string]func(writer http.ResponseWriter, request *http.Request))
	this.patternRoutes = []ServerRoutePattern{}
	this.staticDirs = []ServerStaticDir{}

	// 配置
	this.config = &ServerConfig{}
	this.config.Load()

	if this.singleInstance {
		// 执行参数
		this.execArgs()

		// 检查PID
		this.checkPid()

		// 记录PID
		this.writePid()
	}
}

// Start 启动服务
func (this *Server) Start() {
	if !this.config.Http.On && !this.config.Https.On && len(this.config.Http.Listen) == 0 && len(this.config.Https.Listen) == 0 {
		this.StartOn("0.0.0.0:8888")
	} else {
		this.StartOn("")
	}
}

// StartOn 在某个地址上启动服务
func (this *Server) StartOn(address string) {
	var serverMux = http.NewServeMux()

	// Functions
	beforeStartOnce.Do(func() {
		locker := sync.Mutex{}
		if len(beforeStartFunctions) > 0 {
			for _, fn := range beforeStartFunctions {
				locker.Lock()
				fn(this)
				locker.Unlock()
			}
		}
	})

	// 静态资源目录
	for _, staticDir := range this.staticDirs {
		var staticDirCopy = staticDir
		var prefix = staticDirCopy.prefix
		if len(prefix) > 0 {
			if !strings.HasPrefix(prefix, "/") {
				prefix = "/" + prefix
			}
			if !strings.HasSuffix(prefix, "/") {
				prefix += "/"
			}

			serverMux.HandleFunc(prefix, func(writer http.ResponseWriter, request *http.Request) {
				writer = newResponseWriter(writer)

				// 输出日志
				if this.accessLog {
					defer this.logWriter.Print(time.Now(), writer.(*responseWriter), request)
				}

				_, isText := this.outputMimeType(writer, request.URL.Path)
				if isText {
					newWriter, err := newGzipWriter(writer, 5)
					if err == nil {
						writer = newWriter
						defer func() {
							_ = newWriter.Close()
						}()
					}
				}

				http.StripPrefix(strings.TrimSuffix(prefix, "/"), http.FileServer(http.Dir(staticDirCopy.dir+"/"))).ServeHTTP(writer, request)
			})
		}
	}

	// 加载和动作一致的静态资源
	serverMux.HandleFunc("/_/", func(writer http.ResponseWriter, request *http.Request) {
		writer = newResponseWriter(writer)

		// 输出日志
		if this.accessLog {
			defer this.logWriter.Print(time.Now(), writer.(*responseWriter), request)
		}

		var ext = strings.ToLower(filepath.Ext(request.URL.Path))
		if stringutil.Contains([]string{".html", ""}, ext) || strings.HasPrefix(filepath.Base(request.URL.Path), ".") { // 禁止访问html文件、目录、隐藏文件（.xxx）
			http.Error(writer, "No permission to view page", http.StatusForbidden)
			return
		}
		request.URL.Path = strings.TrimPrefix(request.URL.Path, "/_/")

		_, isText := this.outputMimeType(writer, request.URL.Path)
		if isText {
			newWriter, err := newGzipWriter(writer, 5)
			if err == nil {
				writer = newWriter
				defer func() {
					_ = newWriter.Close()
				}()
			}
		}

		http.FileServer(http.Dir(Tea.ViewsDir())).ServeHTTP(writer, request)
	})

	// 请求处理函数
	var moduleReg, err = stringutil.RegexpCompile("^/+@([\\w-]+)(/.*)$")
	if err != nil {
		panic(err)
		return
	}

	serverMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer = newResponseWriter(writer)

		// 输出日志
		if this.accessLog {
			defer this.logWriter.Print(time.Now(), writer.(*responseWriter), request)
		}

		// call before request callback
		if this.beforeRequestFunc != nil {
			if !this.beforeRequestFunc(writer, request) {
				return
			}
		}

		var requestPath = request.URL.Path

		// 模块
		var parsedResult = moduleReg.FindAllStringSubmatch(requestPath, -1)
		var module = ""
		if len(parsedResult) > 0 {
			module = parsedResult[0][1]
			requestPath = parsedResult[0][2]
		}

		// 路由key
		var key string
		if len(module) > 0 {
			key = "/" + module + requestPath + "__MELOY__" + request.Method
		} else {
			key = requestPath + "__MELOY__" + request.Method
		}

		runFunc, found := this.directRoutes[key]
		if found {
			runFunc(writer, request)
			return
		}

		// 支持 * 路由
		if len(module) > 0 {
			key = "/" + module + "*" + "__MELOY__" + request.Method
		} else {
			key = "*" + "__MELOY__" + request.Method
		}
		runFunc, found = this.directRoutes[key]
		if found {
			runFunc(writer, request)
			return
		}

		// 支持 * 方法 （代表所有请求方法）
		key = requestPath + "__MELOY__*"
		runFunc, found = this.directRoutes[key]
		if found {
			runFunc(writer, request)
			return
		}

		// 查找pattern
		if len(this.patternRoutes) > 0 {
			for _, route := range this.patternRoutes {
				if route.method != "*" && route.method != request.Method {
					continue
				}

				if route.reg.MatchString(requestPath) {
					matches := route.reg.FindStringSubmatch(request.URL.Path)

					values := url.Values{}
					for index, name := range route.names {
						values.Add(name, matches[index+1])
					}

					if len(values) > 0 {
						if len(request.URL.RawQuery) == 0 {
							request.URL.RawQuery = values.Encode()
						} else {
							request.URL.RawQuery += "&" + values.Encode()
						}
					}

					route.runFunc(writer, request)
					return
				}
			}
		}

		// 试图读取静态文件
		publicFilePath := Tea.PublicFile(requestPath)
		stat, err := os.Stat(publicFilePath)
		if err == nil && !stat.IsDir() {
			_, isText := this.outputMimeType(writer, requestPath)

			// 试着压缩
			if isText {
				newWriter, err := newGzipWriter(writer, 5)
				if err == nil {
					writer = newWriter
					defer func() {
						_ = newWriter.Close()
					}()
				}
			}

			http.FileServer(http.Dir(Tea.PublicDir())).ServeHTTP(writer, request)
			return
		}

		// 处理404的情况
		this.config.processError(request, writer, http.StatusNotFound, "404 page not found")
		//http.Error(writer, "404 page not found", http.StatusNotFound)
		return

	})

	// 如果没有指定地址，则从配置中加载
	if len(address) == 0 {
		// http
		if this.config.Http.On {
			for _, addr := range this.config.Http.Listen {
				logs.Println("start http server on", addr)

				server := &http.Server{
					Addr:              addr,
					Handler:           serverMux,
					IdleTimeout:       2 * time.Minute,
					ErrorLog:          this.internalErrorLogger,
					ReadHeaderTimeout: this.readHeaderTimeout,
					ReadTimeout:       this.readTimeout,
				}

				if this.connState != nil {
					server.ConnState = this.connState
				}

				this.httpServerLocker.Lock()
				this.httpServers = append(this.httpServers, server)
				this.httpServerLocker.Unlock()

				go func() {
					err := server.ListenAndServe()
					if err != nil {
						logs.Error(err)
					}
				}()
			}
		}

		// https
		if this.config.Https.On {
			for _, addr := range this.config.Https.Listen {
				logs.Println("start ssl server on", addr)

				server := &http.Server{
					Addr:              addr,
					Handler:           serverMux,
					IdleTimeout:       2 * time.Minute,
					ErrorLog:          this.internalErrorLogger,
					ReadHeaderTimeout: this.readHeaderTimeout,
					ReadTimeout:       this.readTimeout,
				}

				if this.connState != nil {
					server.ConnState = this.connState
				}

				this.httpServerLocker.Lock()
				this.httpServers = append(this.httpServers, server)
				this.httpServerLocker.Unlock()

				go func() {
					certFile := this.config.Https.Cert
					if !filepath.IsAbs(certFile) {
						certFile = Tea.Root + Tea.DS + certFile
					}

					keyFile := this.config.Https.Key
					if !filepath.IsAbs(keyFile) {
						keyFile = Tea.Root + Tea.DS + keyFile
					}

					err := server.ListenAndServeTLS(certFile, keyFile)
					if err != nil {
						logs.Error(errors.New("tls: " + err.Error()))
					}
				}()
			}
		}
	}

	// 默认地址
	if len(address) > 0 {
		logs.Println("start server on", address)
		go func() {
			err := http.ListenAndServe(address, serverMux)
			if err != nil {
				logs.Errorf(err.Error())
				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	// 日志
	if !logs.HasWriter() {
		this.LogWriter(&DefaultLogWriter{})
	}

	// 启动任务管理器
	//tasks.Start(runtime.NumCPU() * 4)

	// 等待
	select {}
}

func (this *Server) Stop() {
	// stop servers
	this.httpServerLocker.Lock()
	for _, server := range this.httpServers {
		_ = server.Close()
	}
	this.httpServerLocker.Unlock()

	// call stop Functions
	beforeStopOnce.Do(func() {
		locker := sync.Mutex{}
		if len(beforeStopFunctions) > 0 {
			for _, fn := range beforeStopFunctions {
				locker.Lock()
				fn(this)
				locker.Unlock()
			}
		}
	})
}

// BeforeRequest 请求前回调
func (this *Server) BeforeRequest(fn func(writer http.ResponseWriter, req *http.Request) bool) *Server {
	this.beforeRequestFunc = fn
	return this
}

// 配置路由
func (this *Server) router(pattern string, method string, actionPtr any) {
	this.routerLocker.Lock()
	defer this.routerLocker.Unlock()

	pattern = this.lastPrefix + pattern

	if reflect.TypeOf(actionPtr).Kind() == reflect.Func { // 函数
		// do nothing
	} else { // struct对象或指针
		if reflect.TypeOf(actionPtr).Kind().String() != "ptr" {
			actionPtr = reflect.New(reflect.TypeOf(actionPtr)).Interface()
		}

		actionPtr.(actions.ActionWrapper).Object().Module = this.lastModule
	}

	method = strings.ToUpper(method)

	// 是否包含匹配参数 :paramName(pattern)
	var reg = actionPathRegexp
	var matches = reg.FindAllStringSubmatch(pattern, 10)
	var names = []string{}
	for _, match := range matches {
		names = append(names, match[1])
		if len(match[3]) > 0 {
			pattern = strings.Replace(pattern, match[0], match[3], 10)
		} else {
			pattern = strings.Replace(pattern, match[0], "([^/]+)", 10)
		}
	}

	var err error
	reg, err = regexp.Compile("^" + pattern + "$")
	if err == nil && len(names) > 0 {
		var routePattern = ServerRoutePattern{
			module:  this.lastModule,
			reg:     reg,
			names:   names,
			method:  method,
			runFunc: this.buildHandle(actionPtr),
		}
		this.patternRoutes = append(this.patternRoutes, routePattern)
		return
	}

	// 正常的路由
	var key string
	if len(this.lastModule) > 0 {
		key = "/" + this.lastModule + pattern + "__MELOY__" + method
	} else {
		key = pattern + "__MELOY__" + method
	}
	this.directRoutes[key] = this.buildHandle(actionPtr)
}

func (this *Server) buildHandle(actionPtr any) func(writer http.ResponseWriter, request *http.Request) {
	// 是否为函数
	if reflect.TypeOf(actionPtr).Kind() == reflect.Func {
		{
			f, ok := actionPtr.(func(request *http.Request, writer http.ResponseWriter))
			if ok {
				return func(writer http.ResponseWriter, request *http.Request) {
					f(request, writer)
				}
			}
		}

		{
			f, ok := actionPtr.(func(writer http.ResponseWriter, request *http.Request))
			if ok {
				return func(writer http.ResponseWriter, request *http.Request) {
					f(writer, request)
				}
			}
		}

		{
			f, ok := actionPtr.(func(request *http.Request))
			if ok {
				return func(writer http.ResponseWriter, request *http.Request) {
					f(request)
				}
			}
		}

		{
			f, ok := actionPtr.(func(writer http.ResponseWriter))
			if ok {
				return func(writer http.ResponseWriter, request *http.Request) {
					f(writer)
				}
			}
		}

		{
			f, ok := actionPtr.(func())
			if ok {
				return func(writer http.ResponseWriter, request *http.Request) {
					f()
				}
			}
		}

		panic("invalid handle function")

		return nil
	}

	actionWrapper, ok := actionPtr.(actions.ActionWrapper)
	if !ok {
		logs.Errorf("actionPtr should be pointer")
		return func(writer http.ResponseWriter, request *http.Request) {

		}
	}

	var spec = actions.NewActionSpec(actionPtr.(actions.ActionWrapper))
	spec.Module = this.lastModule

	var helpers = append([]any{}, this.lastHelpers...)
	var data = actions.Data{}
	if this.lastData != nil {
		for k, v := range this.lastData {
			data[k] = v
		}
	}

	return func(writer http.ResponseWriter, request *http.Request) {
		// URI Query
		var params = actions.Params{}
		for key, values := range request.URL.Query() {
			params[key] = values
		}

		// POST参数
		if request.Method == "POST" {
			var maxSize = int64(this.config.MaxSize())
			if maxSize <= 0 {
				maxSize = 32 << 20
			}
			err := request.ParseMultipartForm(maxSize)
			if err != nil {
				err = request.ParseForm()
				if err != nil {
					logs.Error(err)
				}
			}
			for key, values := range request.Form {
				params[key] = values
			}
		}

		var actionObject = actionWrapper.Object()
		actionObject.SetMaxSize(this.config.MaxSize())
		actionObject.SetSessionManager(this.sessionManager)
		actionObject.SetSessionCookieName(this.sessionCookieName)

		actions.RunAction(actionPtr, spec, request, writer, params, helpers, data)
	}
}

// Module 设置模块定义开始
func (this *Server) Module(module string) *Server {
	this.lastModule = module
	return this
}

// EndModule 设置模块定义结束
func (this *Server) EndModule() *Server {
	this.lastModule = ""
	return this
}

// Prefix 设置URL前缀
func (this *Server) Prefix(prefix string) *Server {
	this.lastPrefix = prefix
	return this
}

// EndPrefix 结束前缀定义
func (this *Server) EndPrefix() *Server {
	this.lastPrefix = ""
	return this
}

// Data 设置变量
func (this *Server) Data(name string, value any) *Server {
	if this.lastData == nil {
		this.lastData = actions.Data{}
	}
	this.lastData[name] = value
	return this
}

// EndData 结束设置变量
func (this *Server) EndData() *Server {
	this.lastData = nil
	return this
}

// Helper 定义助手
func (this *Server) Helper(helper any) *Server {
	if helper == nil {
		logs.Error(errors.New("you try to add a nil helper"))
		return this
	}

	t := reflect.TypeOf(helper).String()

	// 同种类型的Helper只加入一次
	typeStrings := []string{}
	for _, h := range this.lastHelpers {
		if h == nil {
			continue
		}
		t1 := reflect.TypeOf(h).String()
		typeStrings = append(typeStrings, t1)
	}

	if lists.Contains(typeStrings, t) {
		return this
	}

	this.lastHelpers = append(this.lastHelpers, helper)
	return this
}

// EndHelpers 结束助手定义
func (this *Server) EndHelpers() *Server {
	this.lastHelpers = []any{}
	return this
}

// EndAll 结束所有定义
func (this *Server) EndAll() *Server {
	this.EndPrefix()
	this.EndModule()
	this.EndHelpers()
	this.EndData()
	return this
}

// Get 设置 GET 方法路由映射
func (this *Server) Get(path string, actionPtr any) *Server {
	this.router(path, "get", actionPtr)
	return this
}

// Post 设置 POST 方法路由映射
func (this *Server) Post(path string, actionPtr any) *Server {
	this.router(path, "post", actionPtr)
	return this
}

// GetPost 设置 GET 和 POST 方法路由映射
func (this *Server) GetPost(path string, actionPtr any) *Server {
	return this.Any([]string{"get", "post"}, path, actionPtr)
}

// Head 设置 HEAD 方法路由映射
func (this *Server) Head(path string, actionPtr any) *Server {
	this.router(path, "head", actionPtr)
	return this
}

// Delete 设置 DELETE 方法路由映射
func (this *Server) Delete(path string, actionPtr any) *Server {
	this.router(path, "delete", actionPtr)
	return this
}

// Purge 设置 PURGE 方法路由映射
func (this *Server) Purge(path string, actionPtr any) *Server {
	this.router(path, "purge", actionPtr)
	return this
}

// Put 设置 PUT 方法路由映射
func (this *Server) Put(path string, actionPtr any) *Server {
	this.router(path, "put", actionPtr)
	return this
}

// Options 设置 OPTIONS 方法路由映射
func (this *Server) Options(path string, actionPtr any) *Server {
	this.router(path, "options", actionPtr)
	return this
}

// Trace 设置 TRACE 方法路由映射
func (this *Server) Trace(path string, actionPtr any) *Server {
	this.router(path, "trace", actionPtr)
	return this
}

// Connect 设置 CONNECT 方法路由映射
func (this *Server) Connect(path string, actionPtr any) *Server {
	this.router(path, "connect", actionPtr)
	return this
}

// Any 设置一组方法路由映射
func (this *Server) Any(methods []string, path string, actionPtr any) *Server {
	for _, method := range methods {
		this.router(path, method, actionPtr)
	}
	return this
}

// All 将所有方法映射到路由
func (this *Server) All(path string, actionPtr any) *Server {
	this.router(path, "*", actionPtr)
	return this
}

// Static 添加静态目录
func (this *Server) Static(prefix string, dir string) *Server {
	this.staticDirs = append(this.staticDirs, ServerStaticDir{
		prefix: prefix,
		dir:    dir,
	})
	return this
}

// ConnState 连接状态
func (this *Server) ConnState(connState func(conn net.Conn, state http.ConnState)) *Server {
	this.connState = connState
	return this
}

// Session 设置SESSION管理器
func (this *Server) Session(sessionManager any, cookieName string) *Server {
	this.sessionManager = sessionManager
	this.sessionCookieName = cookieName
	return this
}

// LogWriter 设置日志writer
func (this *Server) LogWriter(logWriter LogWriter) *Server {
	if this.logWriter != nil {
		this.logWriter.Close()
	}
	logWriter.Init()
	this.logWriter = logWriter
	logs.SetWriter(logWriter)
	return this
}

// AccessLog 设置是否打印访问日志
func (this *Server) AccessLog(bool bool) *Server {
	this.accessLog = bool
	return this
}

func (this *Server) InternalErrorLogger(errorLogger *log.Logger) *Server {
	this.internalErrorLogger = errorLogger
	return this
}

func (this *Server) ReadHeaderTimeout(timeout time.Duration) *Server {
	this.readHeaderTimeout = timeout
	return this
}

func (this *Server) ReadTimeout(timeout time.Duration) *Server {
	this.readTimeout = timeout
	return this
}

func (this *Server) outputMimeType(writer http.ResponseWriter, path string) (mimeType string, isText bool) {
	ext := filepath.Ext(path)
	if len(ext) == 0 {
		return
	}

	mimeType = mime.TypeByExtension(ext)
	if len(mimeType) == 0 {
		return
	}

	index := strings.Index(mimeType, ";")
	if index > -1 {
		mimeType = mimeType[:index]
	}

	_, found := textMimeMap[mimeType]
	if !found {
		writer.Header().Set("Content-Type", mimeType)
		return
	}

	isText = true
	writer.Header().Set("Content-Type", mimeType+"; charset="+this.config.Charset)
	return
}

// 执行参数，如果找到可执行的参数，则返回 true
func (this *Server) execArgs() {
	if len(os.Args) <= 1 {
		return
	}

	cmd := os.Args[0]
	arg := os.Args[1]
	if arg == "start" { // 启动
		process := processes.NewProcess(cmd)
		_ = process.StartBackground()
		log.Println("started in background")
		os.Exit(0)
	} else if arg == "stop" { // 停止
		defer os.Exit(0)

		pid := this.findPid()
		if pid == 0 {
			log.Println("server not started")
			return
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			log.Println("server not started")
			return
		}

		this.writeNewPid(0)

		_ = process.Kill()

		log.Println("kill pid", pid)

		return
	} else if arg == "restart" { // 重启
		pid := this.findPid()
		if pid > 0 {
			process, err := os.FindProcess(pid)
			if err == nil {
				_ = process.Kill()
			}
		}

		process := processes.NewProcess(cmd)
		_ = process.StartBackground()
		log.Println("started in background")
		os.Exit(0)
	}

	return
}

// 查找PID
func (this *Server) findPid() int {
	pidFile := files.NewFile(Tea.Root + "/bin/pid")
	if !pidFile.IsFile() {
		return 0
	}

	pidString, err := pidFile.ReadAllString()
	if err != nil {
		return 0
	}

	pid := types.Int(pidString)
	if pid <= 0 {
		return 0
	}

	return pid
}

// 检查PID
func (this *Server) checkPid() {
	pid := this.findPid()
	if pid == 0 {
		return
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	logs.Println("kill previous process", pid)
	_ = process.Kill()

	// 等待sock关闭
	time.Sleep(100 * time.Millisecond)
}

// 写入当前程序的PID，以便后续的管理
func (this *Server) writePid() {
	pid := os.Getpid()
	pidDir := files.NewFile(Tea.Root + "/bin")
	if !pidDir.IsDir() {
		err := pidDir.Mkdir()
		if err == nil {
			pidFile := pidDir.Child("pid")
			_ = pidFile.WriteFormat("%d", pid)
		}
	} else {
		pidFile := pidDir.Child("pid")
		_ = pidFile.WriteFormat("%d", pid)
	}
}

// 写入新的PID
func (this *Server) writeNewPid(pid int) {
	pidDir := files.NewFile(Tea.Root + "/bin")
	if !pidDir.IsDir() {
		err := pidDir.Mkdir()
		if err == nil {
			pidFile := pidDir.Child("pid")
			_ = pidFile.WriteFormat("%d", pid)
		}
	} else {
		pidFile := pidDir.Child("pid")
		_ = pidFile.WriteFormat("%d", pid)
	}
}
