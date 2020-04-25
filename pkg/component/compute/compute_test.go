package compute

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ihcsim/pulumi-azure/v2/pkg/mock"
	"github.com/ihcsim/pulumi-azure/v2/pkg/test"
	"github.com/pulumi/pulumi/sdk/go/pulumi"
	"github.com/pulumi/pulumi/sdk/go/pulumi/config"
)

func TestReconcile(t *testing.T) {
	if err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, test.ConfigNamespace)

		resourceGroup, err := test.MockResourceGroup(ctx)
		if err != nil {
			return err
		}

		appSecGroups, err := test.MockApplicationSecurityGroup(ctx)
		if err != nil {
			return err
		}

		virtualNetworks, err := test.MockVirtualNetworks(ctx)
		if err != nil {
			return err
		}

		virtualMachines, err := Reconcile(ctx, cfg, appSecGroups, resourceGroup, virtualNetworks, test.Tags)
		if err != nil {
			return err
		}

		virtualMachine, exists := virtualMachines[test.VirtualMachineInstanceName]
		if !exists {
			return fmt.Errorf("missing virtual machine: %s", test.VirtualMachineName)
		}

		var wg sync.WaitGroup
		wg.Add(1)
		pulumi.All(
			virtualMachine.AvailabilitySetId,
			virtualMachine.Location,
			virtualMachine.Name,
			virtualMachine.OsProfile,
			virtualMachine.OsProfileLinuxConfig,
			virtualMachine.PrimaryNetworkInterfaceId,
			virtualMachine.ResourceGroupName,
			virtualMachine.StorageImageReference,
			virtualMachine.StorageOsDisk,
			virtualMachine.VmSize).ApplyT(func(actuals []interface{}) error {

			defer wg.Done()

			if actual := actuals[0].(string); actual != test.AvailabilitySetName+"_id" {
				t.Errorf("mismatch availability sets. expected: %s, actual: %s", test.AvailabilitySetName, actual)
			}

			if actual := actuals[1].(string); actual != test.Location {
				t.Errorf("mismatch locations. expected: %s, actual: %s", test.Location, actual)
			}

			if actual := actuals[2].(string); actual != test.VirtualMachineInstanceName {
				t.Errorf("mismatch virtual machine name. expected: %s, actual: %s", test.VirtualMachineInstanceName, actual)
			}

			if actual := actuals[5].(*string); *actual != test.NetworkInterfaceName+"_id" {
				t.Errorf("mismatch network interface. expected: %s, actual: %s", test.NetworkInterfaceName+"_id", *actual)
			}

			if actual := actuals[6].(string); actual != test.ResourceGroupName {
				t.Errorf("mismatch resource group. expected: %s, actual: %s", test.ResourceGroupName, actual)
			}

			if actual := actuals[9].(string); actual != test.VirtualMachineSize {
				t.Errorf("mismatch vm size. expected: %s, actual: %s", test.VirtualMachineSize, actual)
			}

			return nil
		})

		wg.Wait()
		return nil
	}, mock.WithCustomMocks(test.Project, test.Stack, test.Config, mock.Mocks(0))); err != nil {
		t.Error(err)
	}
}
