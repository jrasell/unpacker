package main

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jrasell/unpacker/builtin/amazon"
	"github.com/jrasell/unpacker/helper/diff"
)

var (
	dryrun   bool
	region   string
	tagKey   string
	tagValue string
	vrsn     bool

	instKill []string
	saveSg   []string
	saveKp   []string
)

func init() {
	flag.BoolVar(&dryrun, "dryrun", false, "perform a dryrun")
	flag.BoolVar(&vrsn, "version", false, "print version and exit")
	flag.StringVar(&region, "region", "", "region to connect to (required)")
	flag.StringVar(&tagKey, "tag_key", "", "tag key assositated to packer builds (required)")
	flag.StringVar(&tagValue, "tag_value", "", "tag value assositated to packer builds (required)")

	flag.Usage = func() {
		flag.PrintDefaults()
	}

	flag.Parse()
}

func main() {

	if vrsn {
		fmt.Printf("unpacker version %s\n", Version)
		return
	}

	if tagKey == "" || tagValue == "" || region == "" {
		flag.Usage()
		return
	}

	sess, _ := session.NewSession()
	conf := aws.Config{Region: aws.String(region)}
	svc := ec2.New(sess, &conf)

	allIn := amazon.GetPackerInstances(svc, tagKey, tagValue)
	allKp := amazon.GetPackerKeyPairs(svc)
	allSg := amazon.GetPackerSecurityGroups(svc)

	for _, i := range allIn.ToSave {
		saveSg = append(saveSg, i.Sg)
		saveKp = append(saveKp, i.Kp)
	}

	for _, i := range allIn.ToKill {
		instKill = append(instKill, i.ID)
	}

	fmt.Printf("Found %v instances to terminate\n", len(instKill))

	sg := diff.SliceDiff(saveSg, allSg)
	fmt.Printf("Found %v security groups to delete\n", len(sg))

	kp := diff.SliceDiff(saveKp, allKp)
	fmt.Printf("Found %v key pairs to delete\n", len(kp))

	if dryrun == true {
		dryRun(instKill, kp, sg)
		fmt.Println("Unpacker DryRun completed successfully")
		return
	}

	if len(instKill) > 0 {
		_, err := amazon.TerminatePackerInstances(svc, instKill)

		if err != nil {
			println(err)
			return
		}
	}

	for _, k := range allKp {
		amazon.DeletePackerKeyPair(svc, k)
	}
	for _, s := range allSg {
		amazon.DeletePackerSecurityGroup(svc, s)
	}
	fmt.Println("Unpacker completed successfully")
}

func dryRun(in, kp, sg []string) {

	for _, i := range in {
		fmt.Printf("Terminating instance %v - DRYRUN\n", i)
	}

	for _, k := range kp {
		fmt.Printf("Deleting key pair %v - DRYRUN\n", k)
	}

	for _, s := range sg {
		fmt.Printf("Deleting security group %v - DRYRUN\n", s)
	}
}
