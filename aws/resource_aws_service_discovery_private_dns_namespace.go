package aws

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
)

func resourceAwsServiceDiscoveryPrivateDnsNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceDiscoveryPrivateDnsNamespaceCreate,
		Read:   resourceAwsServiceDiscoveryPrivateDnsNamespaceRead,
		Update: resourceAwsServiceDiscoveryPrivateDnsNamespaceUpdate,
		Delete: resourceAwsServiceDiscoveryPrivateDnsNamespaceDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected NAMESPACE_ID:VPC_ID", d.Id())
				}
				d.SetId(idParts[0])
				d.Set("vpc", idParts[1])
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateServiceDiscoveryNamespaceName,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"vpc": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags": tagsSchema(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosted_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsServiceDiscoveryPrivateDnsNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	name := d.Get("name").(string)
	// The CreatorRequestId has a limit of 64 bytes
	var requestId string
	if len(name) > (64 - resource.UniqueIDSuffixLength) {
		requestId = resource.PrefixedUniqueId(name[0:(64 - resource.UniqueIDSuffixLength - 1)])
	} else {
		requestId = resource.PrefixedUniqueId(name)
	}

	input := &servicediscovery.CreatePrivateDnsNamespaceInput{
		Name:             aws.String(name),
		Tags:             keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().ServicediscoveryTags(),
		Vpc:              aws.String(d.Get("vpc").(string)),
		CreatorRequestId: aws.String(requestId),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreatePrivateDnsNamespace(input)

	if err != nil {
		return fmt.Errorf("error creating Service Discovery Private DNS Namespace (%s): %w", name, err)
	}

	if output == nil || output.OperationId == nil {
		return fmt.Errorf("error creating Service Discovery Private DNS Namespace (%s): creation response missing Operation ID", name)
	}

	operationOutput, err := waiter.OperationSuccess(conn, aws.StringValue(output.OperationId))

	if err != nil {
		return fmt.Errorf("error waiting for Service Discovery Private DNS Namespace (%s) creation: %w", name, err)
	}

	if operationOutput == nil || operationOutput.Operation == nil {
		return fmt.Errorf("error creating Service Discovery Private DNS Namespace (%s): operation response missing Operation information", name)
	}

	namespaceID, ok := operationOutput.Operation.Targets[servicediscovery.OperationTargetTypeNamespace]

	if !ok {
		return fmt.Errorf("error creating Service Discovery Private DNS Namespace (%s): operation response missing Namespace ID", name)
	}

	d.SetId(aws.StringValue(namespaceID))

	return resourceAwsServiceDiscoveryPrivateDnsNamespaceRead(d, meta)
}

func resourceAwsServiceDiscoveryPrivateDnsNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	input := &servicediscovery.GetNamespaceInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetNamespace(input)
	if err != nil {
		if isAWSErr(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
			d.SetId("")
			return nil
		}
		return err
	}

	arn := aws.StringValue(resp.Namespace.Arn)
	d.Set("description", resp.Namespace.Description)
	d.Set("arn", arn)
	d.Set("name", resp.Namespace.Name)
	if resp.Namespace.Properties != nil {
		d.Set("hosted_zone", resp.Namespace.Properties.DnsProperties.HostedZoneId)
	}

	tags, err := keyvaluetags.ServicediscoveryListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsServiceDiscoveryPrivateDnsNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.ServicediscoveryUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Service Discovery Private DNS Namespace (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsServiceDiscoveryHttpNamespaceRead(d, meta)
}

func resourceAwsServiceDiscoveryPrivateDnsNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.DeleteNamespaceInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.DeleteNamespace(input)

	if err != nil {
		return fmt.Errorf("error deleting Service Discovery Private DNS Namespace (%s): %w", d.Id(), err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := waiter.OperationSuccess(conn, aws.StringValue(output.OperationId)); err != nil {
			return fmt.Errorf("error waiting for Service Discovery Private DNS Namespace (%s) deletion: %w", d.Id(), err)
		}
	}

	return nil
}
