package main

import "github.com/resulknad/networks2tester/test"
import "regexp"
import "strconv"
import "bufio"
import "log"
import "os"

func ParseTopoFile(f string) *test.Test {
	t := test.NewTest()

	rNodeAdd := regexp.MustCompile(`node[ ]+add[ ]+([^\s]*)[ ]+[a-z]+[ ]+([0-9.]*)/([0-9]*)(?::([0-9]+))?`)
	rInterfaces := regexp.MustCompile(`([0-9.]*)/([0-9]*)(?::([0-9]+))?`)

	rLinkAdd := regexp.MustCompile(`link[ ]+add[ ]+([0-9.]*)[ ]+([0-9.]*)`)

	file, err := os.Open(f)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
		line := scanner.Text()
		if rNodeAdd.MatchString(line) {
			
			matches := rNodeAdd.FindStringSubmatch(line)
			rt := t.AddCustomRouter(matches[1])
			interfaces := rInterfaces.FindAllStringSubmatch(line, -1) 
			for _,match := range interfaces {
				ip := test.Ip2int(match[1])
				maskBits,_ := strconv.ParseUint(match[2], 10, 32)
				var mask uint32
				mask = (0xFFFFFFFF)<<(32-maskBits)
				cost,_ := strconv.Atoi(match[3])
				if match[3] == "" {
					cost = 1
				}
				
				subnet := t.GetOrCreateSubnet(ip, mask)
				intrf := t.GetOrCreateInterface(rt, ip, mask, subnet)
				t.SetInterfaceCost(intrf, float64(cost))
			}
		}

		if rLinkAdd.MatchString(line) {
			match := rLinkAdd.FindStringSubmatch(line)
			intrfAIP := test.Ip2int(match[1])
			intrfBIP := test.Ip2int(match[2])
			t.LinkInterfaces(t.GetInterfaceByIP(intrfAIP), t.GetInterfaceByIP(intrfBIP))
		}
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

	return t
}
