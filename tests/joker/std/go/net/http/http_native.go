// Auto-generated by gostd2joker at (omitted for testing), do not edit!!

package http

import (
	_http "net/http"
	. "github.com/candid82/joker/core"
)

// func error(w ABEND884(unrecognized type ResponseWriter at: tests/big/src/net/http/server.go:1973:14), error string, code int) Object {
// 	_http.Error(w, error, code)
// 	...ABEND675: TODO...
// }

// func fileServer(root ABEND884(unrecognized type FileSystem at: tests/big/src/net/http/fs.go:713:22)) Object {
// 	return _http.FileServer(root)
// }

// func get(url string) Object {
// 	resp, err := _http.Get(url)
// 	_res := EmptyVector
// 	var _obj_map1 Object
// 	if resp != nil {
// 		_map1 := EmptyArrayMap()
// 		_map1.Add(MakeKeyword("Status"), MakeString((*resp).Status))
// 		_map1.Add(MakeKeyword("StatusCode"), MakeInt(int((*resp).StatusCode)))
// 		_map1.Add(MakeKeyword("Proto"), MakeString((*resp).Proto))
// 		_map1.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*resp).ProtoMajor)))
// 		_map1.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*resp).ProtoMinor)))
// 		_map1.Add(MakeKeyword("Header"), (*resp).Header)
// 		_map1.Add(MakeKeyword("Body"), (*resp).Body)
// 		_map1.Add(MakeKeyword("ContentLength"), MakeInt(int((*resp).ContentLength)))
// 		_vec2 := EmptyVector
// 		for _, _elem2 := range (*resp).TransferEncoding {
// 			_vec2 = _vec2.Conjoin(MakeString(_elem2))
// 		}
// 		_map1.Add(MakeKeyword("TransferEncoding"), _vec2)
// 		_map1.Add(MakeKeyword("Close"), MakeBool((*resp).Close))
// 		_map1.Add(MakeKeyword("Uncompressed"), MakeBool((*resp).Uncompressed))
// 		_map1.Add(MakeKeyword("Trailer"), (*resp).Trailer)
// 		var _obj_map3 Object
// 		if (*resp).Request != nil {
// 			_map3 := EmptyArrayMap()
// 			_map3.Add(MakeKeyword("Method"), MakeString((*(*resp).Request).Method))
// 			_map3.Add(MakeKeyword("URL"), (*(*(*resp).Request).URL))
// 			_map3.Add(MakeKeyword("Proto"), MakeString((*(*resp).Request).Proto))
// 			_map3.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*(*resp).Request).ProtoMajor)))
// 			_map3.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*(*resp).Request).ProtoMinor)))
// 			_map3.Add(MakeKeyword("Header"), (*(*resp).Request).Header)
// 			_map3.Add(MakeKeyword("Body"), (*(*resp).Request).Body)
// 			_map3.Add(MakeKeyword("GetBody"), (*(*resp).Request).GetBody)
// 			_map3.Add(MakeKeyword("ContentLength"), MakeInt(int((*(*resp).Request).ContentLength)))
// 			_vec4 := EmptyVector
// 			for _, _elem4 := range (*(*resp).Request).TransferEncoding {
// 				_vec4 = _vec4.Conjoin(MakeString(_elem4))
// 			}
// 			_map3.Add(MakeKeyword("TransferEncoding"), _vec4)
// 			_map3.Add(MakeKeyword("Close"), MakeBool((*(*resp).Request).Close))
// 			_map3.Add(MakeKeyword("Host"), MakeString((*(*resp).Request).Host))
// 			_map3.Add(MakeKeyword("Form"), (*(*resp).Request).Form)
// 			_map3.Add(MakeKeyword("PostForm"), (*(*resp).Request).PostForm)
// 			_map3.Add(MakeKeyword("MultipartForm"), (*(*(*resp).Request).MultipartForm))
// 			_map3.Add(MakeKeyword("Trailer"), (*(*resp).Request).Trailer)
// 			_map3.Add(MakeKeyword("RemoteAddr"), MakeString((*(*resp).Request).RemoteAddr))
// 			_map3.Add(MakeKeyword("RequestURI"), MakeString((*(*resp).Request).RequestURI))
// 			_map3.Add(MakeKeyword("TLS"), (*(*(*resp).Request).TLS))
// 			_map3.Add(MakeKeyword("Cancel"), (*(*resp).Request).Cancel)
// 			_map3.Add(MakeKeyword("Response"), )
// 			_obj_map3 = Object(_map3)
// 		} else {
// 			_obj_map3 = NIL
// 		}
// 		_map1.Add(MakeKeyword("Request"), _obj_map3)
// 		_map1.Add(MakeKeyword("TLS"), (*(*resp).TLS))
// 		_obj_map1 = Object(_map1)
// 	} else {
// 		_obj_map1 = NIL
// 	}
// 	_res = _res.Conjoin(_obj_map1)
// 	_res = _res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeError(err) } }())
// 	return _res
// }

// func handle(pattern string, handler ABEND884(unrecognized type Handler at: tests/big/src/net/http/server.go:2401:37)) Object {
// 	_http.Handle(pattern, handler)
// 	...ABEND675: TODO...
// }

// func handleFunc(pattern string, handler ABEND882(unrecognized Expr type *ast.FuncType at: tests/big/src/net/http/server.go:2406:41)) Object {
// 	_http.HandleFunc(pattern, handler)
// 	...ABEND675: TODO...
// }

// func head(url string) Object {
// 	resp, err := _http.Head(url)
// 	_res := EmptyVector
// 	var _obj_map1 Object
// 	if resp != nil {
// 		_map1 := EmptyArrayMap()
// 		_map1.Add(MakeKeyword("Status"), MakeString((*resp).Status))
// 		_map1.Add(MakeKeyword("StatusCode"), MakeInt(int((*resp).StatusCode)))
// 		_map1.Add(MakeKeyword("Proto"), MakeString((*resp).Proto))
// 		_map1.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*resp).ProtoMajor)))
// 		_map1.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*resp).ProtoMinor)))
// 		_map1.Add(MakeKeyword("Header"), (*resp).Header)
// 		_map1.Add(MakeKeyword("Body"), (*resp).Body)
// 		_map1.Add(MakeKeyword("ContentLength"), MakeInt(int((*resp).ContentLength)))
// 		_vec2 := EmptyVector
// 		for _, _elem2 := range (*resp).TransferEncoding {
// 			_vec2 = _vec2.Conjoin(MakeString(_elem2))
// 		}
// 		_map1.Add(MakeKeyword("TransferEncoding"), _vec2)
// 		_map1.Add(MakeKeyword("Close"), MakeBool((*resp).Close))
// 		_map1.Add(MakeKeyword("Uncompressed"), MakeBool((*resp).Uncompressed))
// 		_map1.Add(MakeKeyword("Trailer"), (*resp).Trailer)
// 		var _obj_map3 Object
// 		if (*resp).Request != nil {
// 			_map3 := EmptyArrayMap()
// 			_map3.Add(MakeKeyword("Method"), MakeString((*(*resp).Request).Method))
// 			_map3.Add(MakeKeyword("URL"), (*(*(*resp).Request).URL))
// 			_map3.Add(MakeKeyword("Proto"), MakeString((*(*resp).Request).Proto))
// 			_map3.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*(*resp).Request).ProtoMajor)))
// 			_map3.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*(*resp).Request).ProtoMinor)))
// 			_map3.Add(MakeKeyword("Header"), (*(*resp).Request).Header)
// 			_map3.Add(MakeKeyword("Body"), (*(*resp).Request).Body)
// 			_map3.Add(MakeKeyword("GetBody"), (*(*resp).Request).GetBody)
// 			_map3.Add(MakeKeyword("ContentLength"), MakeInt(int((*(*resp).Request).ContentLength)))
// 			_vec4 := EmptyVector
// 			for _, _elem4 := range (*(*resp).Request).TransferEncoding {
// 				_vec4 = _vec4.Conjoin(MakeString(_elem4))
// 			}
// 			_map3.Add(MakeKeyword("TransferEncoding"), _vec4)
// 			_map3.Add(MakeKeyword("Close"), MakeBool((*(*resp).Request).Close))
// 			_map3.Add(MakeKeyword("Host"), MakeString((*(*resp).Request).Host))
// 			_map3.Add(MakeKeyword("Form"), (*(*resp).Request).Form)
// 			_map3.Add(MakeKeyword("PostForm"), (*(*resp).Request).PostForm)
// 			_map3.Add(MakeKeyword("MultipartForm"), (*(*(*resp).Request).MultipartForm))
// 			_map3.Add(MakeKeyword("Trailer"), (*(*resp).Request).Trailer)
// 			_map3.Add(MakeKeyword("RemoteAddr"), MakeString((*(*resp).Request).RemoteAddr))
// 			_map3.Add(MakeKeyword("RequestURI"), MakeString((*(*resp).Request).RequestURI))
// 			_map3.Add(MakeKeyword("TLS"), (*(*(*resp).Request).TLS))
// 			_map3.Add(MakeKeyword("Cancel"), (*(*resp).Request).Cancel)
// 			_map3.Add(MakeKeyword("Response"), )
// 			_obj_map3 = Object(_map3)
// 		} else {
// 			_obj_map3 = NIL
// 		}
// 		_map1.Add(MakeKeyword("Request"), _obj_map3)
// 		_map1.Add(MakeKeyword("TLS"), (*(*resp).TLS))
// 		_obj_map1 = Object(_map1)
// 	} else {
// 		_obj_map1 = NIL
// 	}
// 	_res = _res.Conjoin(_obj_map1)
// 	_res = _res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeError(err) } }())
// 	return _res
// }

// func maxBytesReader(w ABEND884(unrecognized type ResponseWriter at: tests/big/src/net/http/request.go:1056:23), r ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/request.go:1056:41), n int64) Object {
// 	return _http.MaxBytesReader(w, r, n)
// }

// func newFileTransport(fs ABEND884(unrecognized type FileSystem at: tests/big/src/net/http/filetransport.go:30:26)) Object {
// 	return _http.NewFileTransport(fs)
// }

// func newRequest(method string, url string, body ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/request.go:792:42)) Object {
// 	_res1, _res2 := _http.NewRequest(method, url, body)
// 	_res := EmptyVector
// 	var _obj_map1 Object
// 	if _res1 != nil {
// 		_map1 := EmptyArrayMap()
// 		_map1.Add(MakeKeyword("Method"), MakeString((*_res1).Method))
// 		_map1.Add(MakeKeyword("URL"), (*(*_res1).URL))
// 		_map1.Add(MakeKeyword("Proto"), MakeString((*_res1).Proto))
// 		_map1.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*_res1).ProtoMajor)))
// 		_map1.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*_res1).ProtoMinor)))
// 		_map1.Add(MakeKeyword("Header"), (*_res1).Header)
// 		_map1.Add(MakeKeyword("Body"), (*_res1).Body)
// 		_map1.Add(MakeKeyword("GetBody"), (*_res1).GetBody)
// 		_map1.Add(MakeKeyword("ContentLength"), MakeInt(int((*_res1).ContentLength)))
// 		_vec2 := EmptyVector
// 		for _, _elem2 := range (*_res1).TransferEncoding {
// 			_vec2 = _vec2.Conjoin(MakeString(_elem2))
// 		}
// 		_map1.Add(MakeKeyword("TransferEncoding"), _vec2)
// 		_map1.Add(MakeKeyword("Close"), MakeBool((*_res1).Close))
// 		_map1.Add(MakeKeyword("Host"), MakeString((*_res1).Host))
// 		_map1.Add(MakeKeyword("Form"), (*_res1).Form)
// 		_map1.Add(MakeKeyword("PostForm"), (*_res1).PostForm)
// 		_map1.Add(MakeKeyword("MultipartForm"), (*(*_res1).MultipartForm))
// 		_map1.Add(MakeKeyword("Trailer"), (*_res1).Trailer)
// 		_map1.Add(MakeKeyword("RemoteAddr"), MakeString((*_res1).RemoteAddr))
// 		_map1.Add(MakeKeyword("RequestURI"), MakeString((*_res1).RequestURI))
// 		_map1.Add(MakeKeyword("TLS"), (*(*_res1).TLS))
// 		_map1.Add(MakeKeyword("Cancel"), (*_res1).Cancel)
// 		var _obj_map3 Object
// 		if (*_res1).Response != nil {
// 			_map3 := EmptyArrayMap()
// 			_map3.Add(MakeKeyword("Status"), MakeString((*(*_res1).Response).Status))
// 			_map3.Add(MakeKeyword("StatusCode"), MakeInt(int((*(*_res1).Response).StatusCode)))
// 			_map3.Add(MakeKeyword("Proto"), MakeString((*(*_res1).Response).Proto))
// 			_map3.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*(*_res1).Response).ProtoMajor)))
// 			_map3.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*(*_res1).Response).ProtoMinor)))
// 			_map3.Add(MakeKeyword("Header"), (*(*_res1).Response).Header)
// 			_map3.Add(MakeKeyword("Body"), (*(*_res1).Response).Body)
// 			_map3.Add(MakeKeyword("ContentLength"), MakeInt(int((*(*_res1).Response).ContentLength)))
// 			_vec4 := EmptyVector
// 			for _, _elem4 := range (*(*_res1).Response).TransferEncoding {
// 				_vec4 = _vec4.Conjoin(MakeString(_elem4))
// 			}
// 			_map3.Add(MakeKeyword("TransferEncoding"), _vec4)
// 			_map3.Add(MakeKeyword("Close"), MakeBool((*(*_res1).Response).Close))
// 			_map3.Add(MakeKeyword("Uncompressed"), MakeBool((*(*_res1).Response).Uncompressed))
// 			_map3.Add(MakeKeyword("Trailer"), (*(*_res1).Response).Trailer)
// 			_map3.Add(MakeKeyword("Request"), )
// 			_map3.Add(MakeKeyword("TLS"), (*(*(*_res1).Response).TLS))
// 			_obj_map3 = Object(_map3)
// 		} else {
// 			_obj_map3 = NIL
// 		}
// 		_map1.Add(MakeKeyword("Response"), _obj_map3)
// 		_obj_map1 = Object(_map1)
// 	} else {
// 		_obj_map1 = NIL
// 	}
// 	_res = _res.Conjoin(_obj_map1)
// 	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
// 	return _res
// }

// func newServeMux() Object {
// 	return _http.NewServeMux()
// 	ABEND124(no public information returned)
// }

// func notFound(w ABEND884(unrecognized type ResponseWriter at: tests/big/src/net/http/server.go:1981:17), r ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/server.go:1981:35)) Object {
// 	_http.NotFound(w, r)
// 	...ABEND675: TODO...
// }

// func notFoundHandler() Object {
// 	return _http.NotFoundHandler()
// }

func parseHTTPVersion(vers string) Object {
	major, minor, ok := _http.ParseHTTPVersion(vers)
	_res := EmptyVector
	_res = _res.Conjoin(MakeInt(int(major)))
	_res = _res.Conjoin(MakeInt(int(minor)))
	_res = _res.Conjoin(MakeBool(ok))
	return _res
}

// func parseTime(text string) Object {
// 	t, err := _http.ParseTime(text)
// 	_res := EmptyVector
// 	_res = _res.Conjoin(t)
// 	_res = _res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeError(err) } }())
// 	return _res
// }

// func post(url string, contentType string, body ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/client.go:748:41)) Object {
// 	resp, err := _http.Post(url, contentType, body)
// 	_res := EmptyVector
// 	var _obj_map1 Object
// 	if resp != nil {
// 		_map1 := EmptyArrayMap()
// 		_map1.Add(MakeKeyword("Status"), MakeString((*resp).Status))
// 		_map1.Add(MakeKeyword("StatusCode"), MakeInt(int((*resp).StatusCode)))
// 		_map1.Add(MakeKeyword("Proto"), MakeString((*resp).Proto))
// 		_map1.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*resp).ProtoMajor)))
// 		_map1.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*resp).ProtoMinor)))
// 		_map1.Add(MakeKeyword("Header"), (*resp).Header)
// 		_map1.Add(MakeKeyword("Body"), (*resp).Body)
// 		_map1.Add(MakeKeyword("ContentLength"), MakeInt(int((*resp).ContentLength)))
// 		_vec2 := EmptyVector
// 		for _, _elem2 := range (*resp).TransferEncoding {
// 			_vec2 = _vec2.Conjoin(MakeString(_elem2))
// 		}
// 		_map1.Add(MakeKeyword("TransferEncoding"), _vec2)
// 		_map1.Add(MakeKeyword("Close"), MakeBool((*resp).Close))
// 		_map1.Add(MakeKeyword("Uncompressed"), MakeBool((*resp).Uncompressed))
// 		_map1.Add(MakeKeyword("Trailer"), (*resp).Trailer)
// 		var _obj_map3 Object
// 		if (*resp).Request != nil {
// 			_map3 := EmptyArrayMap()
// 			_map3.Add(MakeKeyword("Method"), MakeString((*(*resp).Request).Method))
// 			_map3.Add(MakeKeyword("URL"), (*(*(*resp).Request).URL))
// 			_map3.Add(MakeKeyword("Proto"), MakeString((*(*resp).Request).Proto))
// 			_map3.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*(*resp).Request).ProtoMajor)))
// 			_map3.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*(*resp).Request).ProtoMinor)))
// 			_map3.Add(MakeKeyword("Header"), (*(*resp).Request).Header)
// 			_map3.Add(MakeKeyword("Body"), (*(*resp).Request).Body)
// 			_map3.Add(MakeKeyword("GetBody"), (*(*resp).Request).GetBody)
// 			_map3.Add(MakeKeyword("ContentLength"), MakeInt(int((*(*resp).Request).ContentLength)))
// 			_vec4 := EmptyVector
// 			for _, _elem4 := range (*(*resp).Request).TransferEncoding {
// 				_vec4 = _vec4.Conjoin(MakeString(_elem4))
// 			}
// 			_map3.Add(MakeKeyword("TransferEncoding"), _vec4)
// 			_map3.Add(MakeKeyword("Close"), MakeBool((*(*resp).Request).Close))
// 			_map3.Add(MakeKeyword("Host"), MakeString((*(*resp).Request).Host))
// 			_map3.Add(MakeKeyword("Form"), (*(*resp).Request).Form)
// 			_map3.Add(MakeKeyword("PostForm"), (*(*resp).Request).PostForm)
// 			_map3.Add(MakeKeyword("MultipartForm"), (*(*(*resp).Request).MultipartForm))
// 			_map3.Add(MakeKeyword("Trailer"), (*(*resp).Request).Trailer)
// 			_map3.Add(MakeKeyword("RemoteAddr"), MakeString((*(*resp).Request).RemoteAddr))
// 			_map3.Add(MakeKeyword("RequestURI"), MakeString((*(*resp).Request).RequestURI))
// 			_map3.Add(MakeKeyword("TLS"), (*(*(*resp).Request).TLS))
// 			_map3.Add(MakeKeyword("Cancel"), (*(*resp).Request).Cancel)
// 			_map3.Add(MakeKeyword("Response"), )
// 			_obj_map3 = Object(_map3)
// 		} else {
// 			_obj_map3 = NIL
// 		}
// 		_map1.Add(MakeKeyword("Request"), _obj_map3)
// 		_map1.Add(MakeKeyword("TLS"), (*(*resp).TLS))
// 		_obj_map1 = Object(_map1)
// 	} else {
// 		_obj_map1 = NIL
// 	}
// 	_res = _res.Conjoin(_obj_map1)
// 	_res = _res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeError(err) } }())
// 	return _res
// }

// func postForm(url string, data ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/client.go:785:32)) Object {
// 	resp, err := _http.PostForm(url, data)
// 	_res := EmptyVector
// 	var _obj_map1 Object
// 	if resp != nil {
// 		_map1 := EmptyArrayMap()
// 		_map1.Add(MakeKeyword("Status"), MakeString((*resp).Status))
// 		_map1.Add(MakeKeyword("StatusCode"), MakeInt(int((*resp).StatusCode)))
// 		_map1.Add(MakeKeyword("Proto"), MakeString((*resp).Proto))
// 		_map1.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*resp).ProtoMajor)))
// 		_map1.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*resp).ProtoMinor)))
// 		_map1.Add(MakeKeyword("Header"), (*resp).Header)
// 		_map1.Add(MakeKeyword("Body"), (*resp).Body)
// 		_map1.Add(MakeKeyword("ContentLength"), MakeInt(int((*resp).ContentLength)))
// 		_vec2 := EmptyVector
// 		for _, _elem2 := range (*resp).TransferEncoding {
// 			_vec2 = _vec2.Conjoin(MakeString(_elem2))
// 		}
// 		_map1.Add(MakeKeyword("TransferEncoding"), _vec2)
// 		_map1.Add(MakeKeyword("Close"), MakeBool((*resp).Close))
// 		_map1.Add(MakeKeyword("Uncompressed"), MakeBool((*resp).Uncompressed))
// 		_map1.Add(MakeKeyword("Trailer"), (*resp).Trailer)
// 		var _obj_map3 Object
// 		if (*resp).Request != nil {
// 			_map3 := EmptyArrayMap()
// 			_map3.Add(MakeKeyword("Method"), MakeString((*(*resp).Request).Method))
// 			_map3.Add(MakeKeyword("URL"), (*(*(*resp).Request).URL))
// 			_map3.Add(MakeKeyword("Proto"), MakeString((*(*resp).Request).Proto))
// 			_map3.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*(*resp).Request).ProtoMajor)))
// 			_map3.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*(*resp).Request).ProtoMinor)))
// 			_map3.Add(MakeKeyword("Header"), (*(*resp).Request).Header)
// 			_map3.Add(MakeKeyword("Body"), (*(*resp).Request).Body)
// 			_map3.Add(MakeKeyword("GetBody"), (*(*resp).Request).GetBody)
// 			_map3.Add(MakeKeyword("ContentLength"), MakeInt(int((*(*resp).Request).ContentLength)))
// 			_vec4 := EmptyVector
// 			for _, _elem4 := range (*(*resp).Request).TransferEncoding {
// 				_vec4 = _vec4.Conjoin(MakeString(_elem4))
// 			}
// 			_map3.Add(MakeKeyword("TransferEncoding"), _vec4)
// 			_map3.Add(MakeKeyword("Close"), MakeBool((*(*resp).Request).Close))
// 			_map3.Add(MakeKeyword("Host"), MakeString((*(*resp).Request).Host))
// 			_map3.Add(MakeKeyword("Form"), (*(*resp).Request).Form)
// 			_map3.Add(MakeKeyword("PostForm"), (*(*resp).Request).PostForm)
// 			_map3.Add(MakeKeyword("MultipartForm"), (*(*(*resp).Request).MultipartForm))
// 			_map3.Add(MakeKeyword("Trailer"), (*(*resp).Request).Trailer)
// 			_map3.Add(MakeKeyword("RemoteAddr"), MakeString((*(*resp).Request).RemoteAddr))
// 			_map3.Add(MakeKeyword("RequestURI"), MakeString((*(*resp).Request).RequestURI))
// 			_map3.Add(MakeKeyword("TLS"), (*(*(*resp).Request).TLS))
// 			_map3.Add(MakeKeyword("Cancel"), (*(*resp).Request).Cancel)
// 			_map3.Add(MakeKeyword("Response"), )
// 			_obj_map3 = Object(_map3)
// 		} else {
// 			_obj_map3 = NIL
// 		}
// 		_map1.Add(MakeKeyword("Request"), _obj_map3)
// 		_map1.Add(MakeKeyword("TLS"), (*(*resp).TLS))
// 		_obj_map1 = Object(_map1)
// 	} else {
// 		_obj_map1 = NIL
// 	}
// 	_res = _res.Conjoin(_obj_map1)
// 	_res = _res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeError(err) } }())
// 	return _res
// }

// func proxyFromEnvironment(req ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/transport.go:345:31)) Object {
// 	_res1, _res2 := _http.ProxyFromEnvironment(req)
// 	_res := EmptyVector
// 	_res = _res.Conjoin((*_res1))
// 	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
// 	return _res
// }

// func proxyURL(fixedURL ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/transport.go:351:24)) Object {
// 	return _http.ProxyURL(fixedURL)
// }

// func readRequest(b ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/request.go:942:20)) Object {
// 	_res1, _res2 := _http.ReadRequest(b)
// 	_res := EmptyVector
// 	var _obj_map1 Object
// 	if _res1 != nil {
// 		_map1 := EmptyArrayMap()
// 		_map1.Add(MakeKeyword("Method"), MakeString((*_res1).Method))
// 		_map1.Add(MakeKeyword("URL"), (*(*_res1).URL))
// 		_map1.Add(MakeKeyword("Proto"), MakeString((*_res1).Proto))
// 		_map1.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*_res1).ProtoMajor)))
// 		_map1.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*_res1).ProtoMinor)))
// 		_map1.Add(MakeKeyword("Header"), (*_res1).Header)
// 		_map1.Add(MakeKeyword("Body"), (*_res1).Body)
// 		_map1.Add(MakeKeyword("GetBody"), (*_res1).GetBody)
// 		_map1.Add(MakeKeyword("ContentLength"), MakeInt(int((*_res1).ContentLength)))
// 		_vec2 := EmptyVector
// 		for _, _elem2 := range (*_res1).TransferEncoding {
// 			_vec2 = _vec2.Conjoin(MakeString(_elem2))
// 		}
// 		_map1.Add(MakeKeyword("TransferEncoding"), _vec2)
// 		_map1.Add(MakeKeyword("Close"), MakeBool((*_res1).Close))
// 		_map1.Add(MakeKeyword("Host"), MakeString((*_res1).Host))
// 		_map1.Add(MakeKeyword("Form"), (*_res1).Form)
// 		_map1.Add(MakeKeyword("PostForm"), (*_res1).PostForm)
// 		_map1.Add(MakeKeyword("MultipartForm"), (*(*_res1).MultipartForm))
// 		_map1.Add(MakeKeyword("Trailer"), (*_res1).Trailer)
// 		_map1.Add(MakeKeyword("RemoteAddr"), MakeString((*_res1).RemoteAddr))
// 		_map1.Add(MakeKeyword("RequestURI"), MakeString((*_res1).RequestURI))
// 		_map1.Add(MakeKeyword("TLS"), (*(*_res1).TLS))
// 		_map1.Add(MakeKeyword("Cancel"), (*_res1).Cancel)
// 		var _obj_map3 Object
// 		if (*_res1).Response != nil {
// 			_map3 := EmptyArrayMap()
// 			_map3.Add(MakeKeyword("Status"), MakeString((*(*_res1).Response).Status))
// 			_map3.Add(MakeKeyword("StatusCode"), MakeInt(int((*(*_res1).Response).StatusCode)))
// 			_map3.Add(MakeKeyword("Proto"), MakeString((*(*_res1).Response).Proto))
// 			_map3.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*(*_res1).Response).ProtoMajor)))
// 			_map3.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*(*_res1).Response).ProtoMinor)))
// 			_map3.Add(MakeKeyword("Header"), (*(*_res1).Response).Header)
// 			_map3.Add(MakeKeyword("Body"), (*(*_res1).Response).Body)
// 			_map3.Add(MakeKeyword("ContentLength"), MakeInt(int((*(*_res1).Response).ContentLength)))
// 			_vec4 := EmptyVector
// 			for _, _elem4 := range (*(*_res1).Response).TransferEncoding {
// 				_vec4 = _vec4.Conjoin(MakeString(_elem4))
// 			}
// 			_map3.Add(MakeKeyword("TransferEncoding"), _vec4)
// 			_map3.Add(MakeKeyword("Close"), MakeBool((*(*_res1).Response).Close))
// 			_map3.Add(MakeKeyword("Uncompressed"), MakeBool((*(*_res1).Response).Uncompressed))
// 			_map3.Add(MakeKeyword("Trailer"), (*(*_res1).Response).Trailer)
// 			_map3.Add(MakeKeyword("Request"), )
// 			_map3.Add(MakeKeyword("TLS"), (*(*(*_res1).Response).TLS))
// 			_obj_map3 = Object(_map3)
// 		} else {
// 			_obj_map3 = NIL
// 		}
// 		_map1.Add(MakeKeyword("Response"), _obj_map3)
// 		_obj_map1 = Object(_map1)
// 	} else {
// 		_obj_map1 = NIL
// 	}
// 	_res = _res.Conjoin(_obj_map1)
// 	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
// 	return _res
// }

// func readResponse(r ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/response.go:148:21), req ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/response.go:148:40)) Object {
// 	_res1, _res2 := _http.ReadResponse(r, req)
// 	_res := EmptyVector
// 	var _obj_map1 Object
// 	if _res1 != nil {
// 		_map1 := EmptyArrayMap()
// 		_map1.Add(MakeKeyword("Status"), MakeString((*_res1).Status))
// 		_map1.Add(MakeKeyword("StatusCode"), MakeInt(int((*_res1).StatusCode)))
// 		_map1.Add(MakeKeyword("Proto"), MakeString((*_res1).Proto))
// 		_map1.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*_res1).ProtoMajor)))
// 		_map1.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*_res1).ProtoMinor)))
// 		_map1.Add(MakeKeyword("Header"), (*_res1).Header)
// 		_map1.Add(MakeKeyword("Body"), (*_res1).Body)
// 		_map1.Add(MakeKeyword("ContentLength"), MakeInt(int((*_res1).ContentLength)))
// 		_vec2 := EmptyVector
// 		for _, _elem2 := range (*_res1).TransferEncoding {
// 			_vec2 = _vec2.Conjoin(MakeString(_elem2))
// 		}
// 		_map1.Add(MakeKeyword("TransferEncoding"), _vec2)
// 		_map1.Add(MakeKeyword("Close"), MakeBool((*_res1).Close))
// 		_map1.Add(MakeKeyword("Uncompressed"), MakeBool((*_res1).Uncompressed))
// 		_map1.Add(MakeKeyword("Trailer"), (*_res1).Trailer)
// 		var _obj_map3 Object
// 		if (*_res1).Request != nil {
// 			_map3 := EmptyArrayMap()
// 			_map3.Add(MakeKeyword("Method"), MakeString((*(*_res1).Request).Method))
// 			_map3.Add(MakeKeyword("URL"), (*(*(*_res1).Request).URL))
// 			_map3.Add(MakeKeyword("Proto"), MakeString((*(*_res1).Request).Proto))
// 			_map3.Add(MakeKeyword("ProtoMajor"), MakeInt(int((*(*_res1).Request).ProtoMajor)))
// 			_map3.Add(MakeKeyword("ProtoMinor"), MakeInt(int((*(*_res1).Request).ProtoMinor)))
// 			_map3.Add(MakeKeyword("Header"), (*(*_res1).Request).Header)
// 			_map3.Add(MakeKeyword("Body"), (*(*_res1).Request).Body)
// 			_map3.Add(MakeKeyword("GetBody"), (*(*_res1).Request).GetBody)
// 			_map3.Add(MakeKeyword("ContentLength"), MakeInt(int((*(*_res1).Request).ContentLength)))
// 			_vec4 := EmptyVector
// 			for _, _elem4 := range (*(*_res1).Request).TransferEncoding {
// 				_vec4 = _vec4.Conjoin(MakeString(_elem4))
// 			}
// 			_map3.Add(MakeKeyword("TransferEncoding"), _vec4)
// 			_map3.Add(MakeKeyword("Close"), MakeBool((*(*_res1).Request).Close))
// 			_map3.Add(MakeKeyword("Host"), MakeString((*(*_res1).Request).Host))
// 			_map3.Add(MakeKeyword("Form"), (*(*_res1).Request).Form)
// 			_map3.Add(MakeKeyword("PostForm"), (*(*_res1).Request).PostForm)
// 			_map3.Add(MakeKeyword("MultipartForm"), (*(*(*_res1).Request).MultipartForm))
// 			_map3.Add(MakeKeyword("Trailer"), (*(*_res1).Request).Trailer)
// 			_map3.Add(MakeKeyword("RemoteAddr"), MakeString((*(*_res1).Request).RemoteAddr))
// 			_map3.Add(MakeKeyword("RequestURI"), MakeString((*(*_res1).Request).RequestURI))
// 			_map3.Add(MakeKeyword("TLS"), (*(*(*_res1).Request).TLS))
// 			_map3.Add(MakeKeyword("Cancel"), (*(*_res1).Request).Cancel)
// 			_map3.Add(MakeKeyword("Response"), )
// 			_obj_map3 = Object(_map3)
// 		} else {
// 			_obj_map3 = NIL
// 		}
// 		_map1.Add(MakeKeyword("Request"), _obj_map3)
// 		_map1.Add(MakeKeyword("TLS"), (*(*_res1).TLS))
// 		_obj_map1 = Object(_map1)
// 	} else {
// 		_obj_map1 = NIL
// 	}
// 	_res = _res.Conjoin(_obj_map1)
// 	_res = _res.Conjoin(func () Object { if (_res2) == nil { return NIL } else { return MakeError(_res2) } }())
// 	return _res
// }

// func redirect(w ABEND884(unrecognized type ResponseWriter at: tests/big/src/net/http/server.go:2020:17), r ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/server.go:2020:35), url string, code int) Object {
// 	_http.Redirect(w, r, url, code)
// 	...ABEND675: TODO...
// }

// func redirectHandler(url string, code int) Object {
// 	return _http.RedirectHandler(url, code)
// }

// func serveContent(w ABEND884(unrecognized type ResponseWriter at: tests/big/src/net/http/fs.go:151:21), req ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/fs.go:151:41), name string, modtime ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/fs.go:151:72), content ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/fs.go:151:91)) Object {
// 	_http.ServeContent(w, req, name, modtime, content)
// 	...ABEND675: TODO...
// }

// func serveFile(w ABEND884(unrecognized type ResponseWriter at: tests/big/src/net/http/fs.go:670:18), r ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/fs.go:670:36), name string) Object {
// 	_http.ServeFile(w, r, name)
// 	...ABEND675: TODO...
// }

// func setCookie(w ABEND884(unrecognized type ResponseWriter at: tests/big/src/net/http/cookie.go:157:18), cookie ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/http/cookie.go:157:41)) Object {
// 	_http.SetCookie(w, cookie)
// 	...ABEND675: TODO...
// }

// func stripPrefix(prefix string, h ABEND884(unrecognized type Handler at: tests/big/src/net/http/server.go:1992:35)) Object {
// 	return _http.StripPrefix(prefix, h)
// }

// func timeoutHandler(h ABEND884(unrecognized type Handler at: tests/big/src/net/http/server.go:3106:23), dt ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/http/server.go:3106:35), msg string) Object {
// 	return _http.TimeoutHandler(h, dt, msg)
// }
