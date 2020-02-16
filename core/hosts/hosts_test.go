package hosts

import (
	"io/ioutil"
	"net"
	"os"
	"testing"

	"github.com/shawn1m/overture/core/finder/full"
)

func TestHosts_Find(t *testing.T) {

	hostLinesString := []string{"1.2.3.4 abc.com\n", "::1 abc.com\n", "2.3.4.5 abc.com\n", "::2 abc.com\n",
		"::1 localhost\n", "127.0.0.1 localhost\n"}
	hostsFile, err := generateHostsFile(hostLinesString)
	if err != nil {
		t.Error(err)
	}

	hosts, err := New(hostsFile, &full.Map{DataMap: make(map[string][]string, 100)})
	if err != nil {
		t.Error(err)
	}

	ipv4List, ipv6List := hosts.Find("abc.com")
	if !find(ipv4List, net.ParseIP("1.2.3.4")) {
		t.Error()
	}
	if !find(ipv4List, net.ParseIP("2.3.4.5")) {
		t.Error()
	}
	if !find(ipv6List, net.ParseIP("::1")) {
		t.Error()
	}
	if !find(ipv6List, net.ParseIP("::2")) {
		t.Error()
	}

	ipv4List, ipv6List = hosts.Find("localhost")
	if !find(ipv4List, net.ParseIP("127.0.0.1")) {
		t.Error()
	}
	if !find(ipv6List, net.ParseIP("::1")) {
		t.Error()
	}
}

func generateHostsFile(hostLinesString []string) (string, error) {

	var f *os.File
	f, err := ioutil.TempFile("", "hosts_test")
	if err != nil {
		return "", err
	}
	for _, hostLineString := range hostLinesString {
		if _, err := f.WriteString(hostLineString); err != nil {
			return "", err
		}
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func find(a []net.IP, x net.IP) bool {
	for _, n := range a {
		if x.Equal(n) {
			return true
		}
	}
	return false
}
