package v1

import (
	"testing"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/acceptance/clients"
	"github.com/opentelekomcloud/gophertelekomcloud/acceptance/tools"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/subnets"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v1/vpcs"
	th "github.com/opentelekomcloud/gophertelekomcloud/testhelper"
)

func TestSubnetList(t *testing.T) {
	client, err := clients.NewNetworkV1Client()
	if err != nil {
		t.Fatalf("Unable to create NetworkV1 client : %v", err)
	}
	allPages, err := subnets.List(client, subnets.ListOpts{})
	th.AssertNoErr(t, err)
	tools.PrintResource(t, allPages)
}

func TestSubnetsCRUD(t *testing.T) {
	client, err := clients.NewNetworkV1Client()
	if err != nil {
		t.Fatalf("Unable to create NetworkV1 client: %v", err)
	}

	// Create a subnet
	subnet, err := createSubnetNResources(t, client)
	if err != nil {
		t.Fatalf("Unable to create subnet: %v", err)
	}
	// Delete a subnet
	defer deleteSubnetNResources(t, client, subnet.VPC_ID, subnet.ID)

	tools.PrintResource(t, subnet)

	// wait to be active
	t.Logf("waitting for subnet %s to be active", subnet.ID)
	if err := WaitForSubnetToActive(client, subnet.ID, 60); err != nil {
		t.Fatalf("Error deleting subnet: %v", err)
	}

	// Update a subnet
	newName := tools.RandomString("ACPTTEST-", 8)
	updateOpts := &subnets.UpdateOpts{
		Name: newName,
	}
	t.Logf("Attempting to update name of subnet to %s", newName)
	_, err = subnets.Update(client, subnet.VPC_ID, subnet.ID, updateOpts).Extract()
	if err != nil {
		t.Fatalf("Unable to update subnet: %v", err)
	}

	// Query a subnet
	newSubnet, err := subnets.Get(client, subnet.ID).Extract()
	if err != nil {
		t.Fatalf("Unable to retrieve subnet: %v", err)
	}

	tools.PrintResource(t, newSubnet)
}

func createSubnetNResources(t *testing.T, client *golangsdk.ServiceClient) (*subnets.Subnet, error) {
	vpcName := tools.RandomString("TESTACC-", 8)

	createOpts := vpcs.CreateOpts{
		Name: vpcName,
		CIDR: "192.168.20.0/24",
	}

	t.Logf("Attempting to create vpc: %s", vpcName)

	vpc, err := vpcs.Create(client, createOpts).Extract()
	if err != nil {
		return nil, err
	}
	t.Logf("Created vpc: %s", vpcName)

	subnetName := tools.RandomString("ACPTTEST-", 8)

	createSubnetOpts := subnets.CreateOpts{
		Name:       subnetName,
		CIDR:       "192.168.20.0/24",
		GatewayIP:  "192.168.20.1",
		EnableDHCP: true,
		VPC_ID:     vpc.ID,
	}

	t.Logf("Attempting to create subnet: %s", subnetName)

	subnet, err := subnets.Create(client, createSubnetOpts).Extract()
	if err != nil {
		return subnet, err
	}
	t.Logf("Created subnet: %v", subnet)

	return subnet, nil
}

func deleteSubnetNResources(t *testing.T, client *golangsdk.ServiceClient, vpcID string, id string) {
	t.Logf("Attempting to delete subnet: %s", id)

	err := subnets.Delete(client, vpcID, id).ExtractErr()
	if err != nil {
		t.Fatalf("Error deleting subnet: %v", err)
	}

	t.Logf("waitting for subnet %s to delete", id)
	if err := WaitForSubnetToDelete(client, id, 60); err != nil {
		t.Fatalf("Error deleting subnet: %v", err)
	}

	t.Logf("Deleted subnet: %s", id)
	t.Logf("Attempting to delete vpc: %s", vpcID)

	err = vpcs.Delete(client, vpcID).ExtractErr()
	if err != nil {
		t.Fatalf("Error deleting vpc: %v", err)
	}

	t.Logf("Deleted vpc: %s", vpcID)
}
