// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	BucketResourceAccessIdPartsCount = 2
)

// @SDKResource("aws_lightsail_bucket_resource_access")
func ResourceBucketResourceAccess() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketResourceAccessCreate,
		ReadWithoutTimeout:   resourceBucketResourceAccessRead,
		DeleteWithoutTimeout: resourceBucketResourceAccessDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-z0-9][a-z0-9-]{1,52}[a-z0-9]$`), "Invalid Bucket name. Must match regex: ^[a-z0-9][a-z0-9-]{1,52}[a-z0-9]$"),
			},
			"resource_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`\w[\w\-]*\w`), "Invalid resource name. Must match regex: \\w[\\w\\-]*\\w"),
			},
		},
	}
}

func resourceBucketResourceAccessCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	in := lightsail.SetResourceAccessForBucketInput{
		BucketName:   aws.String(d.Get("bucket_name").(string)),
		ResourceName: aws.String(d.Get("resource_name").(string)),
		Access:       types.ResourceBucketAccessAllow,
	}

	out, err := conn.SetResourceAccessForBucket(ctx, &in)

	if err != nil {
		return create.DiagError(names.Lightsail, string(types.OperationTypeSetResourceAccessForBucket), ResBucketResourceAccess, d.Get("bucket_name").(string), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeSetResourceAccessForBucket, ResBucketResourceAccess, d.Get("bucket_name").(string))

	if diag != nil {
		return diag
	}

	idParts := []string{d.Get("bucket_name").(string), d.Get("resource_name").(string)}
	id, err := flex.FlattenResourceId(idParts, BucketResourceAccessIdPartsCount, false)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionFlatteningResourceId, ResBucketResourceAccess, d.Get("bucket_name").(string), err)
	}

	d.SetId(id)

	return resourceBucketResourceAccessRead(ctx, d, meta)
}

func resourceBucketResourceAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	out, err := FindBucketResourceAccessById(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.Lightsail, create.ErrActionReading, ResBucketResourceAccess, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionReading, ResBucketResourceAccess, d.Id(), err)
	}

	parts, err := flex.ExpandResourceId(d.Id(), BucketResourceAccessIdPartsCount, false)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceId, ResBucketResourceAccess, d.Id(), err)
	}

	d.Set("bucket_name", parts[0])
	d.Set("resource_name", out.Name)

	return nil
}

func resourceBucketResourceAccessDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailClient(ctx)
	parts, err := flex.ExpandResourceId(d.Id(), BucketResourceAccessIdPartsCount, false)

	if err != nil {
		return create.DiagError(names.Lightsail, create.ErrActionExpandingResourceId, ResBucketResourceAccess, d.Id(), err)
	}

	out, err := conn.SetResourceAccessForBucket(ctx, &lightsail.SetResourceAccessForBucketInput{
		BucketName:   aws.String(parts[0]),
		ResourceName: aws.String(parts[1]),
		Access:       types.ResourceBucketAccessDeny,
	})

	if err != nil && errs.IsA[*types.NotFoundException](err) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.Lightsail, string(types.OperationTypeSetResourceAccessForBucket), ResBucketResourceAccess, d.Id(), err)
	}

	diag := expandOperations(ctx, conn, out.Operations, types.OperationTypeSetResourceAccessForBucket, ResBucketResourceAccess, d.Id())

	if diag != nil {
		return diag
	}

	return nil
}

func FindBucketResourceAccessById(ctx context.Context, conn *lightsail.Client, id string) (*types.ResourceReceivingAccess, error) {
	parts, err := flex.ExpandResourceId(id, BucketAccessKeyIdPartsCount, false)

	if err != nil {
		return nil, err
	}

	in := &lightsail.GetBucketsInput{
		BucketName:                aws.String(parts[0]),
		IncludeConnectedResources: aws.Bool(true),
	}

	out, err := conn.GetBuckets(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || len(out.Buckets) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	bucket := out.Buckets[0]
	var entry types.ResourceReceivingAccess
	entryExists := false

	for _, n := range bucket.ResourcesReceivingAccess {
		if parts[1] == aws.ToString(n.Name) {
			entry = n
			entryExists = true
			break
		}
	}

	if !entryExists {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &entry, nil
}
