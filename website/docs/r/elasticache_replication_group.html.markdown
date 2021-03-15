---
subcategory: "ElastiCache"
layout: "aws"
page_title: "AWS: aws_elasticache_replication_group"
description: |-
  Provides an ElastiCache Replication Group resource.
---

# Resource: aws_elasticache_replication_group

Provides an ElastiCache Replication Group resource.
For working with Memcached or single primary Redis instances (Cluster Mode Disabled), see the
[`aws_elasticache_cluster` resource](/docs/providers/aws/r/elasticache_cluster.html).

~> **Note:** When you change an attribute, such as `engine_version`, by
default the ElastiCache API applies it in the next maintenance window. Because
of this, Terraform may report a difference in its planning phase because the
actual modification has not yet taken place. You can use the
`apply_immediately` flag to instruct the service to apply the change
immediately. Using `apply_immediately` can result in a brief downtime as
servers reboots.

~> **Note:** Be aware of the terminology collision around "cluster" for `aws_elasticache_replication_group`. For example, it is possible to create a ["Cluster Mode Disabled [Redis] Cluster"](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Clusters.Create.CON.Redis.html). With "Cluster Mode Enabled", the data will be stored in shards (called "node groups"). See [Redis Cluster Configuration](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/cluster-create-determine-requirements.html#redis-cluster-configuration) for a diagram of the differences. To enable cluster mode, use a parameter group that has cluster mode enabled. The default parameter groups provided by AWS end with ".cluster.on", for example `default.redis6.x.cluster.on`.

## Example Usage

### Redis Cluster Mode Disabled

To create a single shard primary with single read replica:

```hcl
resource "aws_elasticache_replication_group" "example" {
  automatic_failover_enabled    = true
  availability_zones            = ["us-west-2a", "us-west-2b"]
  replication_group_id          = "tf-rep-group-1"
  replication_group_description = "test description"
  node_type                     = "cache.m4.large"
  number_cache_clusters         = 2
  parameter_group_name          = "default.redis3.2"
  port                          = 6379
}
```

You have two options for adjusting the number of replicas:

* Adjusting `number_cache_clusters` directly. This will attempt to automatically add or remove replicas, but provides no granular control (e.g. preferred availability zone, cache cluster ID) for the added or removed replicas. This also currently expects cache cluster IDs in the form of `replication_group_id-00#`.
* Otherwise for fine grained control of the underlying cache clusters, they can be added or removed with the [`aws_elasticache_cluster` resource](/docs/providers/aws/r/elasticache_cluster.html) and its `replication_group_id` attribute. In this situation, you will need to utilize the [lifecycle configuration block](https://www.terraform.io/docs/configuration/meta-arguments/lifecycle.html) with `ignore_changes` to prevent perpetual differences during Terraform plan with the `number_cache_cluster` attribute.

```hcl
resource "aws_elasticache_replication_group" "example" {
  automatic_failover_enabled    = true
  availability_zones            = ["us-west-2a", "us-west-2b"]
  replication_group_id          = "tf-rep-group-1"
  replication_group_description = "test description"
  node_type                     = "cache.m4.large"
  number_cache_clusters         = 2
  parameter_group_name          = "default.redis3.2"
  port                          = 6379

  lifecycle {
    ignore_changes = [number_cache_clusters]
  }
}

resource "aws_elasticache_cluster" "replica" {
  count = 1

  cluster_id           = "tf-rep-group-1-${count.index}"
  replication_group_id = aws_elasticache_replication_group.example.id
}
```

### Redis Cluster Mode Enabled

To create two shards with a primary and a single read replica each:

```hcl
resource "aws_elasticache_replication_group" "baz" {
  replication_group_id          = "tf-redis-cluster"
  replication_group_description = "test description"
  node_type                     = "cache.t2.small"
  port                          = 6379
  parameter_group_name          = "default.redis3.2.cluster.on"
  automatic_failover_enabled    = true

  cluster_mode {
    replicas_per_node_group = 1
    num_node_groups         = 2
  }
}
```

~> **Note:** We currently do not support passing a `primary_cluster_id` in order to create the Replication Group.

~> **Note:** Automatic Failover is unavailable for Redis versions earlier than 2.8.6,
and unavailable on T1 node types. For T2 node types, it is only available on Redis version 3.2.4 or later with cluster mode enabled. See the [High Availability Using Replication Groups](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Replication.html) guide
for full details on using Replication Groups.

## Argument Reference

The following arguments are supported:

* `replication_group_id` – (Required) The replication group identifier. This parameter is stored as a lowercase string.
* `replication_group_description` – (Required) A user-created description for the replication group.
* `number_cache_clusters` - (Optional) The number of cache clusters (primary and replicas) this replication group will have. If Multi-AZ is enabled, the value of this parameter must be at least 2. Updates will occur before other modifications. One of `number_cache_clusters` or `cluster_mode` is required.
* `node_type` - (Required) The compute and memory capacity of the nodes in the node group.
* `automatic_failover_enabled` - (Optional) Specifies whether a read-only replica will be automatically promoted to read/write primary if the existing primary fails. If true, Multi-AZ is enabled for this replication group. If false, Multi-AZ is disabled for this replication group. Must be enabled for Redis (cluster mode enabled) replication groups. Defaults to `false`.
* `auto_minor_version_upgrade` - (Optional) Specifies whether a minor engine upgrades will be applied automatically to the underlying Cache Cluster instances during the maintenance window. This parameter is currently not supported by the AWS API. Defaults to `true`.
* `availability_zones` - (Optional) A list of EC2 availability zones in which the replication group's cache clusters will be created. The order of the availability zones in the list is not important.
* `engine` - (Optional) The name of the cache engine to be used for the clusters in this replication group. e.g. `redis`
* `at_rest_encryption_enabled` - (Optional) Whether to enable encryption at rest.
* `transit_encryption_enabled` - (Optional) Whether to enable encryption in transit.
* `auth_token` - (Optional) The password used to access a password protected server. Can be specified only if `transit_encryption_enabled = true`.
* `kms_key_id` - (Optional) The ARN of the key that you wish to use if encrypting at rest. If not supplied, uses service managed encryption. Can be specified only if `at_rest_encryption_enabled = true`.
* `engine_version` - (Optional) The version number of the cache engine to be used for the cache clusters in this replication group.
* `parameter_group_name` - (Optional) The name of the parameter group to associate with this replication group. If this argument is omitted, the default cache parameter group for the specified engine is used. To enable "cluster mode", i.e. data sharding, use a parameter group that has the parameter `cluster-enabled` set to true.
* `port` – (Optional) The port number on which each of the cache nodes will accept connections. For Memcache the default is 11211, and for Redis the default port is 6379.
* `subnet_group_name` - (Optional) The name of the cache subnet group to be used for the replication group.
* `security_group_names` - (Optional) A list of cache security group names to associate with this replication group.
* `security_group_ids` - (Optional) One or more Amazon VPC security groups associated with this replication group. Use this parameter only when you are creating a replication group in an Amazon Virtual Private Cloud
* `snapshot_arns` – (Optional) A single-element string list containing an
Amazon Resource Name (ARN) of a Redis RDB snapshot file stored in Amazon S3.
Example: `arn:aws:s3:::my_bucket/snapshot1.rdb`
* `snapshot_name` - (Optional) The name of a snapshot from which to restore data into the new node group. Changing the `snapshot_name` forces a new resource.
* `maintenance_window` – (Optional) Specifies the weekly time range for when maintenance
on the cache cluster is performed. The format is `ddd:hh24:mi-ddd:hh24:mi` (24H Clock UTC).
The minimum maintenance window is a 60 minute period. Example: `sun:05:00-sun:09:00`
* `notification_topic_arn` – (Optional) An Amazon Resource Name (ARN) of an
SNS topic to send ElastiCache notifications to. Example:
`arn:aws:sns:us-east-1:012345678999:my_sns_topic`
* `snapshot_window` - (Optional, Redis only) The daily time range (in UTC) during which ElastiCache will
begin taking a daily snapshot of your cache cluster. The minimum snapshot window is a 60 minute period. Example: `05:00-09:00`
* `snapshot_retention_limit` - (Optional, Redis only) The number of days for which ElastiCache will
retain automatic cache cluster snapshots before deleting them. For example, if you set
SnapshotRetentionLimit to 5, then a snapshot that was taken today will be retained for 5 days
before being deleted. If the value of SnapshotRetentionLimit is set to zero (0), backups are turned off.
Please note that setting a `snapshot_retention_limit` is not supported on cache.t1.micro cache nodes
* `apply_immediately` - (Optional) Specifies whether any modifications are applied immediately, or during the next maintenance window. Default is `false`.
* `tags` - (Optional) A map of tags to assign to the resource. Adding tags to this resource will add or overwrite any existing tags on the clusters in the replication group and not to the group itself.
* `final_snapshot_identifier` - (Optional) The name of your final node group (shard) snapshot. ElastiCache creates the snapshot from the primary node in the cluster. If omitted, no final snapshot will be made.
* `cluster_mode` - (Optional) Create a native Redis cluster. `automatic_failover_enabled` must be set to true. Cluster Mode documented below. Only 1 `cluster_mode` block is allowed. One of `number_cache_clusters` or `cluster_mode` is required. Note that configuring this block does not enable cluster mode, i.e. data sharding, this requires using a parameter group that has the parameter `cluster-enabled` set to true.

Cluster Mode (`cluster_mode`) supports the following:

* `replicas_per_node_group` - (Required) Specify the number of replica nodes in each node group. Valid values are 0 to 5. Changing this number will force a new resource.
* `num_node_groups` - (Required) Specify the number of node groups (shards) for this Redis replication group. Changing this number will trigger an online resizing operation before other settings modifications.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The ID of the ElastiCache Replication Group.
* `cluster_enabled` - Indicates if cluster mode is enabled.
* `configuration_endpoint_address` - The address of the replication group configuration endpoint when cluster mode is enabled.
* `primary_endpoint_address` - (Redis only) The address of the endpoint for the primary node in the replication group, if the cluster mode is disabled.
* `reader_endpoint_address` - (Redis only) The address of the endpoint for the reader node in the replication group, if the cluster mode is disabled.
* `member_clusters` - The identifiers of all the nodes that are part of this replication group.

## Timeouts

`aws_elasticache_replication_group` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts)
configuration options:

* `create` - (Default `60m`) How long to wait for a replication group to be created.
* `delete` - (Default `40m`) How long to wait for a replication group to be deleted.
* `update` - (Default `40m`) How long to wait for replication group settings to be updated. This is also separately used for adding/removing replicas and online resize operation completion, if necessary.

## Import

ElastiCache Replication Groups can be imported using the `replication_group_id`, e.g.

```
$ terraform import aws_elasticache_replication_group.my_replication_group replication-group-1
```
