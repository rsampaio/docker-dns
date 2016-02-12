package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
	"github.com/samalba/dockerclient"
)

var docker dockerclient.DockerClient
var dockerBr = flag.String("i", "docker0", "docker bridge")

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	records := make([]dns.RR, 0)
	q := r.Question[0]

	if q.Qtype == dns.TypeA && strings.HasSuffix(q.Name, ".docker.") {
		docker, _ := dockerclient.NewDockerClient("unix:///var/run/docker.sock", &tls.Config{})
		name := strings.SplitN(q.Name, ".", 2)[0]
		containers, err := docker.ListContainers(false, false, fmt.Sprintf("{\"name\":[\"%s\"]}", name))
		if err != nil {
			log.Fatal(err)
		}
		for _, c := range containers {
			info, _ := docker.InspectContainer(c.Id)
			log.Printf("Container %s[%6s] has ip %s\n", name, info.Id, info.NetworkSettings.IPAddress)
			records = append(records,
				&dns.A{
					Hdr: dns.RR_Header{
						Name:   q.Name,
						Rrtype: dns.TypeA,
						Class:  dns.ClassINET,
						Ttl:    60},
					A: net.ParseIP(info.NetworkSettings.IPAddress),
				})
		}
	}

	m.Answer = append(m.Answer, records...)
	defer w.WriteMsg(m)
}

func main() {
	flag.Parse()
	dockerInt, err := net.InterfaceByName(*dockerBr)
	if err != nil {
		panic(err)
	}

	dockerIpNet, err := dockerInt.Addrs()
	if err != nil {
		panic(err)
	}

	dockerIp := strings.Split(dockerIpNet[0].String(), "/")

	server := &dns.Server{Addr: dockerIp[0] + ":53", Net: "udp"}
	dns.HandleFunc(".", handleDnsRequest)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
