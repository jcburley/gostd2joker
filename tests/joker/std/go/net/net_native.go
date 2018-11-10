// Auto-generated by gostd2joker at (omitted for testing), do not edit!!

package net

import (
	"net"
	. "github.com/candid82/joker/core"
)

func cIDRMask(ones int, bits int) Object {
	res := net.CIDRMask(ones, bits)
	vec1 := EmptyVector
	for _, elem1 := range res {
		vec1 = vec1.Conjoin(MakeInt(int(elem1)))
	}
	return vec1
}

// func dial(network string, address string) Object {
// 	res1, res2 := net.Dial(network, address)
// 	res := EmptyVector
// 	res = res.Conjoin(res1)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func dialIP(network string, laddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/iprawsock.go:211:42), raddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/iprawsock.go:211:42)) Object {
// 	res1, res2 := net.DialIP(network, laddr, raddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func dialTCP(network string, laddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/tcpsock.go:206:43), raddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/tcpsock.go:206:43)) Object {
// 	res1, res2 := net.DialTCP(network, laddr, raddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func dialTimeout(network string, address string, timeout ABEND882(unrecognized Expr type *ast.SelectorExpr at: tests/big/src/net/dial.go:313:51)) Object {
// 	res1, res2 := net.DialTimeout(network, address, timeout)
// 	res := EmptyVector
// 	res = res.Conjoin(res1)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func dialUDP(network string, laddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/udpsock.go:205:43), raddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/udpsock.go:205:43)) Object {
// 	res1, res2 := net.DialUDP(network, laddr, raddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func dialUnix(network string, laddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/unixsock.go:200:44), raddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/unixsock.go:200:44)) Object {
// 	res1, res2 := net.DialUnix(network, laddr, raddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func fileConn(f ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/file.go:21:17)) Object {
// 	c, err := net.FileConn(f)
// 	res := EmptyVector
// 	res = res.Conjoin(c)
// 	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
// 	return res
// }

// func fileListener(f ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/file.go:33:21)) Object {
// 	ln, err := net.FileListener(f)
// 	res := EmptyVector
// 	res = res.Conjoin(ln)
// 	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
// 	return res
// }

// func filePacketConn(f ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/file.go:45:23)) Object {
// 	c, err := net.FilePacketConn(f)
// 	res := EmptyVector
// 	res = res.Conjoin(c)
// 	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
// 	return res
// }

func iPv4(a byte, b byte, c byte, d byte) Object {
	res := net.IPv4(a, b, c, d)
	vec1 := EmptyVector
	for _, elem1 := range res {
		vec1 = vec1.Conjoin(MakeInt(int(elem1)))
	}
	return vec1
}

func iPv4Mask(a byte, b byte, c byte, d byte) Object {
	res := net.IPv4Mask(a, b, c, d)
	vec1 := EmptyVector
	for _, elem1 := range res {
		vec1 = vec1.Conjoin(MakeInt(int(elem1)))
	}
	return vec1
}

// func interfaceAddrs() Object {
// 	res1, res2 := net.InterfaceAddrs()
// 	res := EmptyVector
// 	vec1 := EmptyVector
// 	for _, elem1 := range res1 {
// 		vec1 = vec1.Conjoin(elem1)
// 	}
// 	res = res.Conjoin(vec1)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

func interfaceByIndex(index int) Object {
	res1, res2 := net.InterfaceByIndex(index)
	res := EmptyVector
	map1 := EmptyArrayMap()
	map1.Add(MakeKeyword("Index"), MakeInt(int((*res1).Index)))
	map1.Add(MakeKeyword("MTU"), MakeInt(int((*res1).MTU)))
	map1.Add(MakeKeyword("Name"), MakeString((*res1).Name))
	vec2 := EmptyVector
	for _, elem2 := range (*res1).HardwareAddr {
		vec2 = vec2.Conjoin(MakeInt(int(elem2)))
	}
	map1.Add(MakeKeyword("HardwareAddr"), vec2)
	map1.Add(MakeKeyword("Flags"), MakeInt(int((*res1).Flags)))
	res = res.Conjoin(map1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func interfaceByName(name string) Object {
	res1, res2 := net.InterfaceByName(name)
	res := EmptyVector
	map1 := EmptyArrayMap()
	map1.Add(MakeKeyword("Index"), MakeInt(int((*res1).Index)))
	map1.Add(MakeKeyword("MTU"), MakeInt(int((*res1).MTU)))
	map1.Add(MakeKeyword("Name"), MakeString((*res1).Name))
	vec2 := EmptyVector
	for _, elem2 := range (*res1).HardwareAddr {
		vec2 = vec2.Conjoin(MakeInt(int(elem2)))
	}
	map1.Add(MakeKeyword("HardwareAddr"), vec2)
	map1.Add(MakeKeyword("Flags"), MakeInt(int((*res1).Flags)))
	res = res.Conjoin(map1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func interfaces() Object {
	res1, res2 := net.Interfaces()
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range res1 {
		map2 := EmptyArrayMap()
		map2.Add(MakeKeyword("Index"), MakeInt(int(elem1.Index)))
		map2.Add(MakeKeyword("MTU"), MakeInt(int(elem1.MTU)))
		map2.Add(MakeKeyword("Name"), MakeString(elem1.Name))
		vec3 := EmptyVector
		for _, elem3 := range elem1.HardwareAddr {
			vec3 = vec3.Conjoin(MakeInt(int(elem3)))
		}
		map2.Add(MakeKeyword("HardwareAddr"), vec3)
		map2.Add(MakeKeyword("Flags"), MakeInt(int(elem1.Flags)))
		vec1 = vec1.Conjoin(map2)
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

// func listen(network string, address string) Object {
// 	res1, res2 := net.Listen(network, address)
// 	res := EmptyVector
// 	res = res.Conjoin(res1)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func listenIP(network string, laddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/iprawsock.go:230:37)) Object {
// 	res1, res2 := net.ListenIP(network, laddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func listenMulticastUDP(network string, ifi ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/udpsock.go:265:45), gaddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/udpsock.go:265:63)) Object {
// 	res1, res2 := net.ListenMulticastUDP(network, ifi, gaddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func listenPacket(network string, address string) Object {
// 	res1, res2 := net.ListenPacket(network, address)
// 	res := EmptyVector
// 	res = res.Conjoin(res1)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func listenTCP(network string, laddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/tcpsock.go:323:38)) Object {
// 	res1, res2 := net.ListenTCP(network, laddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func listenUDP(network string, laddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/udpsock.go:231:38)) Object {
// 	res1, res2 := net.ListenUDP(network, laddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func listenUnix(network string, laddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/unixsock.go:314:39)) Object {
// 	res1, res2 := net.ListenUnix(network, laddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

// func listenUnixgram(network string, laddr ABEND882(unrecognized Expr type *ast.StarExpr at: tests/big/src/net/unixsock.go:334:43)) Object {
// 	res1, res2 := net.ListenUnixgram(network, laddr)
// 	res := EmptyVector
// 	res = res.Conjoin(NIL)
// 	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
// 	return res
// }

func lookupAddr(addr string) Object {
	names, err := net.LookupAddr(addr)
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range names {
		vec1 = vec1.Conjoin(MakeString(elem1))
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
	return res
}

func lookupCNAME(host string) Object {
	cname, err := net.LookupCNAME(host)
	res := EmptyVector
	res = res.Conjoin(MakeString(cname))
	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
	return res
}

func lookupHost(host string) Object {
	addrs, err := net.LookupHost(host)
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range addrs {
		vec1 = vec1.Conjoin(MakeString(elem1))
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
	return res
}

func lookupIP(host string) Object {
	res1, res2 := net.LookupIP(host)
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range res1 {
		vec2 := EmptyVector
		for _, elem2 := range elem1 {
			vec2 = vec2.Conjoin(MakeInt(int(elem2)))
		}
		vec1 = vec1.Conjoin(vec2)
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func lookupMX(name string) Object {
	res1, res2 := net.LookupMX(name)
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range res1 {
		map2 := EmptyArrayMap()
		map2.Add(MakeKeyword("Host"), MakeString((*elem1).Host))
		map2.Add(MakeKeyword("Pref"), MakeInt(int((*elem1).Pref)))
		vec1 = vec1.Conjoin(map2)
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func lookupNS(name string) Object {
	res1, res2 := net.LookupNS(name)
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range res1 {
		map2 := EmptyArrayMap()
		map2.Add(MakeKeyword("Host"), MakeString((*elem1).Host))
		vec1 = vec1.Conjoin(map2)
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func lookupPort(network string, service string) Object {
	port, err := net.LookupPort(network, service)
	res := EmptyVector
	res = res.Conjoin(MakeInt(int(port)))
	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
	return res
}

func lookupSRV(service string, proto string, name string) Object {
	cname, addrs, err := net.LookupSRV(service, proto, name)
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range addrs {
		map2 := EmptyArrayMap()
		map2.Add(MakeKeyword("Target"), MakeString((*elem1).Target))
		map2.Add(MakeKeyword("Port"), MakeInt(int((*elem1).Port)))
		map2.Add(MakeKeyword("Priority"), MakeInt(int((*elem1).Priority)))
		map2.Add(MakeKeyword("Weight"), MakeInt(int((*elem1).Weight)))
		vec1 = vec1.Conjoin(map2)
	}
	res = res.Conjoin(MakeString(cname))
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
	return res
}

func lookupTXT(name string) Object {
	res1, res2 := net.LookupTXT(name)
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range res1 {
		vec1 = vec1.Conjoin(MakeString(elem1))
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func parseCIDR(s string) Object {
	res1, res2, res3 := net.ParseCIDR(s)
	res := EmptyVector
	map2 := EmptyArrayMap()
	vec3 := EmptyVector
	for _, elem3 := range (*res2).IP {
		vec3 = vec3.Conjoin(MakeInt(int(elem3)))
	}
	map2.Add(MakeKeyword("IP"), vec3)
	vec4 := EmptyVector
	for _, elem4 := range (*res2).Mask {
		vec4 = vec4.Conjoin(MakeInt(int(elem4)))
	}
	map2.Add(MakeKeyword("Mask"), vec4)
	vec1 := EmptyVector
	for _, elem1 := range res1 {
		vec1 = vec1.Conjoin(MakeInt(int(elem1)))
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(map2)
	res = res.Conjoin(func () Object { if (res3) == nil { return NIL } else { return MakeString(res3.Error()) } }())
	return res
}

func parseIP(s string) Object {
	res := net.ParseIP(s)
	vec1 := EmptyVector
	for _, elem1 := range res {
		vec1 = vec1.Conjoin(MakeInt(int(elem1)))
	}
	return vec1
}

func parseMAC(s string) Object {
	hw, err := net.ParseMAC(s)
	res := EmptyVector
	vec1 := EmptyVector
	for _, elem1 := range hw {
		vec1 = vec1.Conjoin(MakeInt(int(elem1)))
	}
	res = res.Conjoin(vec1)
	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
	return res
}

// func pipe() Object {
// 	res1, res2 := net.Pipe()
// 	res := EmptyVector
// 	res = res.Conjoin(res1)
// 	res = res.Conjoin(res2)
// 	return res
// }

func resolveIPAddr(network string, address string) Object {
	res1, res2 := net.ResolveIPAddr(network, address)
	res := EmptyVector
	map1 := EmptyArrayMap()
	vec2 := EmptyVector
	for _, elem2 := range (*res1).IP {
		vec2 = vec2.Conjoin(MakeInt(int(elem2)))
	}
	map1.Add(MakeKeyword("IP"), vec2)
	map1.Add(MakeKeyword("Zone"), MakeString((*res1).Zone))
	res = res.Conjoin(map1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func resolveTCPAddr(network string, address string) Object {
	res1, res2 := net.ResolveTCPAddr(network, address)
	res := EmptyVector
	map1 := EmptyArrayMap()
	vec2 := EmptyVector
	for _, elem2 := range (*res1).IP {
		vec2 = vec2.Conjoin(MakeInt(int(elem2)))
	}
	map1.Add(MakeKeyword("IP"), vec2)
	map1.Add(MakeKeyword("Port"), MakeInt(int((*res1).Port)))
	map1.Add(MakeKeyword("Zone"), MakeString((*res1).Zone))
	res = res.Conjoin(map1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func resolveUDPAddr(network string, address string) Object {
	res1, res2 := net.ResolveUDPAddr(network, address)
	res := EmptyVector
	map1 := EmptyArrayMap()
	vec2 := EmptyVector
	for _, elem2 := range (*res1).IP {
		vec2 = vec2.Conjoin(MakeInt(int(elem2)))
	}
	map1.Add(MakeKeyword("IP"), vec2)
	map1.Add(MakeKeyword("Port"), MakeInt(int((*res1).Port)))
	map1.Add(MakeKeyword("Zone"), MakeString((*res1).Zone))
	res = res.Conjoin(map1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func resolveUnixAddr(network string, address string) Object {
	res1, res2 := net.ResolveUnixAddr(network, address)
	res := EmptyVector
	map1 := EmptyArrayMap()
	map1.Add(MakeKeyword("Name"), MakeString((*res1).Name))
	map1.Add(MakeKeyword("Net"), MakeString((*res1).Net))
	res = res.Conjoin(map1)
	res = res.Conjoin(func () Object { if (res2) == nil { return NIL } else { return MakeString(res2.Error()) } }())
	return res
}

func splitHostPort(hostport string) Object {
	host, port, err := net.SplitHostPort(hostport)
	res := EmptyVector
	res = res.Conjoin(MakeString(host))
	res = res.Conjoin(MakeString(port))
	res = res.Conjoin(func () Object { if (err) == nil { return NIL } else { return MakeString(err.Error()) } }())
	return res
}