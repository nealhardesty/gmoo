package main

import (
	"fmt"

	"context"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func GetHostedZoneID(client *route53.Client, zoneName string) (string, error) {
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

func ChangeRecordSet(client *route53.Client, zoneID, dnsName, address string) (*route53.ChangeResourceRecordSetsOutput, error) {
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
