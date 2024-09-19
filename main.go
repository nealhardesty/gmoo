package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"context"
  "github.com/nealhardesty/gmoo/internal/netutil"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func getHostedZoneID(client *route53.Client, zoneName string) (string, error) {
	input := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(zoneName),
	}

	resp, err := client.ListHostedZonesByName(context.TODO(), input)
	if err != nil {
		return "", err
	}

	for _, zone := range resp.HostedZones {
		if *zone.Name == zoneName {
			return *zone.Id, nil
		}
	}

	return "", fmt.Errorf("hosted zone %s not found", zoneName)
}

func changeRecordSet(client *route53.Client, zoneID, dnsName, address string) (*route53.ChangeResourceRecordSetsOutput, error) {
	dnsType := "A"
	ttl := int64(60)
	input := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(dnsName),
						Type: types.RRType(dnsType),
						ResourceRecords: []types.ResourceRecord{
							{
								Value: aws.String(address),
							},
						},
						TTL: aws.Int64(ttl),
					},
				},
			},
		},
	}

	r, err := client.ChangeResourceRecordSets(context.TODO(), input)
	return r, err
}

func main() {
	help := flag.Bool("help", false, "print help and exit")
	hostname := flag.String("hostname", "", "hostname to upsert, defaults to actual hostname")
	publicIp := flag.String("publicip", "", "public ip to upsert, defaults to actual public ip from icanhazip.com")
	zone := flag.String("zone", "roadwaffle.com", "zone name to upsert")

	flag.Parse()

	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *hostname == "" {
		var err error
		*hostname, err = netutil.GetHostname()
		if err != nil {
			panic(err)
		}
		*hostname = strings.ToLower(*hostname)
	}
	fmt.Printf("hostname=%s\n", *hostname)

	if *publicIp == "" {
		var err error
		*publicIp, err = netutil.GetPublicIP()
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("publicIp=%s\n", *publicIp)

	*zone = strings.ToLower(*zone)
	*zone = fmt.Sprintf("%s.", *zone)
	fmt.Printf("zone=%s\n", *zone)

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(err)
	}

	r53client := route53.NewFromConfig(cfg)

	zoneID, err := getHostedZoneID(r53client, *zone)
	if err != nil {
		panic(err)
	}
	fmt.Printf("zoneID=%s\n", zoneID)

	dnsName := fmt.Sprintf("%s.%s", *hostname, *zone)
	fmt.Printf("dnsName=%s\n", dnsName)

	response, err := changeRecordSet(r53client, zoneID, dnsName, *publicIp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Done.\n%s\n", response.ChangeInfo.Status)

}
