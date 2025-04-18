package util

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type IpSearch struct {
	prefStart [256]uint32
	prefEnd   [256]uint32
	endArr    []uint32
	addrArr   []string
}

// GetExternalIP 获取本服务器的外网IP
func GetExternalIP() (string, error) {
	resp, err := http.Get("https://ipw.cn/api/ip/myip")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	resultBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(resultBytes)), nil
}

// GetClientPublicIP 尽最大努力实现获取客户端公网 IP 的算法。
// 解析 X-Real-IP 和 X-Forwarded-For 以便于反向代理（nginx 或 haproxy）可以正常工作。
func GetClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" {
			return ip
		}
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

// GetIPAddress 通过IP获取地址
func GetIPAddress(ip string) (province string, city string, err error) {
	var resp *http.Response
	resp, err = http.Get(fmt.Sprintf("https://restapi.amap.com/v3/ip?key=7e30415c3e9ce73d93d20189b9539be8&ip=%s", ip))
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = errors.New("查询地址失败！")
		return
	}
	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var resultMap map[string]interface{}
	resultMap, err = JsonToMap(string(data))
	if err != nil {
		return
	}
	provinceObj := resultMap["province"]
	cityObj := resultMap["city"]
	if provinceObj != nil && cityObj != nil {
		var ok bool
		province, ok = provinceObj.(string)
		if !ok {
			return
		}
		city, ok = cityObj.(string)
		if !ok {
			return
		}
		return
	}
	return
}

// GetIntranetIP 获取本机IP
func GetIntranetIP() (ips []string, err error) {
	ips = make([]string, 0)

	ifaces, e := net.Interfaces()
	if e != nil {
		return ips, e
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}

		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}

		// ignore docker and warden bridge
		if strings.HasPrefix(iface.Name, "docker") || strings.HasPrefix(iface.Name, "w-") {
			continue
		}

		addrs, e := iface.Addrs()
		if e != nil {
			return ips, e
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}

			ipStr := ip.String()
			if IsIntranet(ipStr) {
				ips = append(ips, ipStr)
			}
		}
	}

	return ips, nil
}

// IsIntranet IsIntranet
func IsIntranet(ipStr string) bool {
	if strings.HasPrefix(ipStr, "10.") || strings.HasPrefix(ipStr, "192.168.") {
		return true
	}

	if strings.HasPrefix(ipStr, "172.") {
		// 172.16.0.0-172.31.255.255
		arr := strings.Split(ipStr, ".")
		if len(arr) != 4 {
			return false
		}

		second, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			return false
		}

		if second >= 16 && second <= 31 {
			return true
		}
	}

	return false
}

var instance *IpSearch
var once sync.Once

func GetInstance() *IpSearch {
	once.Do(func() {
		instance = &IpSearch{}
		var err error
		//_, filename, _, _ := runtime.Caller(0)
		//dir := filepath.Dir(filename)
		//
		//// 构建数据文件的完整路径
		//datPath := filepath.Join(dir, "")
		//fmt.Println(" path " + datPath)
		instance, err = LoadDat("/mnt/qqzeng-ip-3.0-ultimate.dat")
		if err != nil {
			log.Fatal("the IP Dat loaded failed!")
		}
	})
	return instance
}

func LoadDat(file string) (*IpSearch, error) {
	p := IpSearch{}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(" err ")
		return nil, err
	}

	for k := 0; k < 256; k++ {
		i := k*8 + 4
		p.prefStart[k] = ReadLittleEndian32(data[i], data[i+1], data[i+2], data[i+3])
		p.prefEnd[k] = ReadLittleEndian32(data[i+4], data[i+5], data[i+6], data[i+7])
	}

	RecordSize := int(ReadLittleEndian32(data[0], data[1], data[2], data[3]))

	p.endArr = make([]uint32, RecordSize)
	p.addrArr = make([]string, RecordSize)
	for i := 0; i < RecordSize; i++ {
		j := 2052 + (i * 8)
		endipnum := ReadLittleEndian32(data[j], data[1+j], data[2+j], data[3+j])
		offset := ReadLittleEndian24(data[4+j], data[5+j], data[6+j])
		length := uint32(data[7+j])
		p.endArr[i] = endipnum
		p.addrArr[i] = string(data[offset:int(offset+length)])
	}
	return &p, err

}

func (p *IpSearch) Get(ip string) string {
	ips := strings.Split(ip, ".")
	pref, err := strconv.Atoi(ips[0])
	if err != nil {
		// 处理错误情况
		return "转换失败" // 或其他适当的错误处理
	}

	low := p.prefStart[pref]
	high := p.prefEnd[pref]
	intIP := ip2Long(ip)
	var cur uint32
	if low == high {
		cur = low
	} else {
		cur = p.binarySearch(low, high, intIP)
	}
	return p.addrArr[cur]
}

func (p *IpSearch) GetArea(ip string) string {
	area := "CN"
	if instance == nil {
		return area
	}

	result := p.Get(ip)
	split := strings.Split(result, "|")
	if len(split) > 8 {
		area = split[8]
	}
	return area
}

func (p *IpSearch) binarySearch(low uint32, high uint32, k uint32) uint32 {
	var M uint32 = 0
	for low <= high {
		mid := (low + high) / 2
		endipNum := p.endArr[mid]
		if endipNum >= k {
			M = mid
			if mid == 0 {
				break
			}
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return M
}

func ip2Long(ipstr string) uint32 {
	ip := net.ParseIP(ipstr)
	ip = ip.To4()
	return binary.BigEndian.Uint32(ip)
}

func ReadLittleEndian32(a, b, c, d byte) uint32 {
	return (uint32(a) & 0xFF) | ((uint32(b) << 8) & 0xFF00) | ((uint32(c) << 16) & 0xFF0000) | ((uint32(d) << 24) & 0xFF000000)
}

func ReadLittleEndian24(a, b, c byte) uint32 {
	return (uint32(a) & 0xFF) | ((uint32(b) << 8) & 0xFF00) | ((uint32(c) << 16) & 0xFF0000)
}
