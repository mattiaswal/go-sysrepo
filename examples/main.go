package main

import (
	"fmt"
	"log"

	"github.com/mattiaswal/go-libyang/libyang"
	"github.com/mattiaswal/go-sysrepo/sysrepo"
)

func main() {
	conn, err := sysrepo.Connect(sysrepo.ConnDefault)
	if err != nil {
		log.Fatalf("Failed to connect to sysrepo: %v", err)
	}
	defer conn.Close()

	sess, err := conn.SessionStart(sysrepo.DSOperational)
	if err != nil {
		log.Fatalf("Failed to start session: %v", err)
	}
	defer sess.Close()

	getData(sess)
	getItem(sess)
}

func getData(sess *sysrepo.Session) {
	fmt.Println("Example: Get data from sysrepo")

	path := "/ietf-system:system-state/ntp/sources"

	node, _ := sess.GetData(path, 0, 0, 0)
	defer node.Free()

	fmt.Print(node.Print(libyang.DataFormatJSON))
	fmt.Println("=============================")

	child := node.Child()
	sources := child.ChildByName("sources")

	for a := sources.Child(); a.Ptr != nil; a = a.Next() {
		child := a.Child()
		for b := child.FirstSibling(); b.Ptr != nil; b = b.Next() {
			fmt.Printf("%s:%s\n", b.Name(), b.Value())
		}
		fmt.Println("--------------------------")
	}
	fmt.Println("=============================")
}

func getItem(sess *sysrepo.Session) {
	fmt.Println("GetItem from sysrepo")
	hostname, err := sess.GetItem("/ietf-system:system/hostname")
	if err != nil {
		println("Error Getting hostname")
	} else {
		fmt.Println("Hostname: " + hostname)
	}
}
