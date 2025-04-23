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
	source := sources.ChildByName("source")
	for n := source.Child(); n.Ptr != nil; n = n.Next() {
		fmt.Printf("Found matching node: %s, value: %s\n", n.Name(), n.Value())
	}

}
