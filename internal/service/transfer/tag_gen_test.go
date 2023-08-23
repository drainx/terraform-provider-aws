// Code generated by internal/generate/tagresource/main.go; DO NOT EDIT.

package transfer_test

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftransfer "github.com/hashicorp/terraform-provider-aws/internal/service/transfer"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func testAccCheckTagDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_transfer_tag" {
				continue
			}

			identifier, key, err := tftags.GetResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tftransfer.GetTag(ctx, conn, identifier, key)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("%s resource (%s) tag (%s) still exists", transfer.ServiceID, identifier, key)
		}

		return nil
	}
}

func testAccCheckTagExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("%s: missing resource ID", resourceName)
		}

		identifier, key, err := tftags.GetResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TransferConn(ctx)

		_, err = tftransfer.GetTag(ctx, conn, identifier, key)

		return err
	}
}
