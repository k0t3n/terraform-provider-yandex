package yandex

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/redis/v1"
)

const (
	yandexMDBRedisClusterDefaultTimeout = 15 * time.Minute
	yandexMDBRedisClusterUpdateTimeout  = 60 * time.Minute
	defaultMDBPageSize                  = 1000
)

func resourceYandexMDBRedisCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBRedisClusterCreate,
		Read:   resourceYandexMDBRedisClusterRead,
		Update: resourceYandexMDBRedisClusterUpdate,
		Delete: resourceYandexMDBRedisClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBRedisClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexMDBRedisClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBRedisClusterDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseRedisEnv),
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"timeout": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"maxmemory_policy": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"notify_keyspace_events": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"slowlog_log_slower_than": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"slowlog_max_len": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"databases": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"version": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"resources": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disk_size": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"disk_type_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"host": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:     schema.TypeString,
							Required: true,
						},
						"shard_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"sharded": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"tls_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
			"maintenance_window": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"ANYTIME", "WEEKLY"}, false),
							Required:     true,
						},
						"day": {
							Type:         schema.TypeString,
							ValidateFunc: validateParsableValue(parseRedisWeekDay),
							Optional:     true,
						},
						"hour": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(1, 24),
							Optional:     true,
						},
					},
				},
			},
		},
	}
}

func resourceYandexMDBRedisClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareCreateRedisRequest(d, config)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Redis().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Redis Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get redis create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*redis.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create Redis Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Redis Cluster creation failed: %s", err)
	}

	mw, err := expandRedisMaintenanceWindow(d)
	if err != nil {
		return err
	}
	if mw != nil {
		err = updateRedisMaintenanceWindow(ctx, config, d, mw)
		if err != nil {
			return err
		}
	}

	return resourceYandexMDBRedisClusterRead(d, meta)
}

func prepareCreateRedisRequest(d *schema.ResourceData, meta *Config) (*redis.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, fmt.Errorf("Error while expanding labels on Redis Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating Redis Cluster: %s", err)
	}

	hosts, err := expandRedisHosts(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding hosts on Redis Cluster create: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseRedisEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating Redis Cluster: %s", err)
	}

	conf, version, err := expandRedisConfig(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding config while creating Redis Cluster: %s", err)
	}

	resources, err := expandRedisResources(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding resources on Redis Cluster create: %s", err)
	}

	configSpec := &redis.ConfigSpec{
		RedisSpec: *conf,
		Resources: resources,
		Version:   version,
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	req := redis.CreateClusterRequest{
		FolderId:         folderID,
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		NetworkId:        d.Get("network_id").(string),
		Environment:      env,
		ConfigSpec:       configSpec,
		HostSpecs:        hosts,
		Labels:           labels,
		Sharded:          d.Get("sharded").(bool),
		TlsEnabled:       &wrappers.BoolValue{Value: d.Get("tls_enabled").(bool)},
		SecurityGroupIds: securityGroupIds,
	}
	return &req, nil
}

func resourceYandexMDBRedisClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Redis().Cluster().Get(ctx, &redis.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	hosts, err := listRedisHosts(ctx, config, d)
	if err != nil {
		return err
	}

	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)
	d.Set("sharded", cluster.Sharded)
	d.Set("tls_enabled", cluster.TlsEnabled)

	resources, err := flattenRedisResources(cluster.Config.Resources)
	if err != nil {
		return err
	}

	conf := extractRedisConfig(cluster.Config)
	password := ""
	if v, ok := d.GetOk("config.0.password"); ok {
		password = v.(string)
	}

	err = d.Set("config", []map[string]interface{}{
		{
			"timeout":                 conf.timeout,
			"maxmemory_policy":        conf.maxmemoryPolicy,
			"notify_keyspace_events":  conf.notifyKeyspaceEvents,
			"slowlog_log_slower_than": conf.slowlogLogSlowerThan,
			"slowlog_max_len":         conf.slowlogMaxLen,
			"databases":               conf.databases,
			"version":                 conf.version,
			"password":                password,
		},
	})
	if err != nil {
		return err
	}

	if err := d.Set("resources", resources); err != nil {
		return err
	}

	// Do not change the state if only order of hosts differs.
	dHosts, err := expandRedisHosts(d)
	if err != nil {
		return err
	}

	sortRedisHosts(hosts, dHosts)

	hs, err := flattenRedisHosts(hosts)
	if err != nil {
		return err
	}

	if err := d.Set("host", hs); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	mw := flattenRedisMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return err
	}

	return d.Set("labels", cluster.Labels)
}

func resourceYandexMDBRedisClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	if d.HasChange("resources.0.disk_type_id") {
		return fmt.Errorf("Changing disk_type_id is not supported for Redis Cluster. Id: %v", d.Id())
	}

	if d.HasChange("name") || d.HasChange("labels") || d.HasChange("description") || d.HasChange("resources") || d.HasChange("config") || d.HasChange("security_group_ids") {
		if err := updateRedisClusterParams(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		if err := updateRedisClusterHosts(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)
	return resourceYandexMDBRedisClusterRead(d, meta)
}

func updateRedisClusterParams(d *schema.ResourceData, meta interface{}) error {
	req := &redis.UpdateClusterRequest{
		ClusterId: d.Id(),
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{},
		},
	}
	onDone := []func(){}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")

		onDone = append(onDone, func() {
			d.SetPartial("name")
		})
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")

		onDone = append(onDone, func() {
			d.SetPartial("labels")
		})
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")

		onDone = append(onDone, func() {
			d.SetPartial("description")
		})
	}

	if d.HasChange("resources") {
		res, err := expandRedisResources(d)
		if err != nil {
			return err
		}

		if req.ConfigSpec == nil {
			req.ConfigSpec = &redis.ConfigSpec{}
		}

		req.ConfigSpec.Resources = res
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.resources")

		onDone = append(onDone, func() {
			d.SetPartial("resources")
		})
	}

	if d.HasChange("config") {
		if d.HasChange("config.0.version") {
			return fmt.Errorf("Version update for Redis is not supported")
		}
		conf, version, err := expandRedisConfig(d)
		if err != nil {
			return err
		}

		if req.ConfigSpec == nil {
			req.ConfigSpec = &redis.ConfigSpec{}
		}

		req.ConfigSpec.RedisSpec = *conf
		switch version {
		case "5.0":
			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.redis_config_5_0")
		case "6.0":
			req.UpdateMask.Paths = append(req.UpdateMask.Paths, "config_spec.redis_config_6_0")
		}

		onDone = append(onDone, func() {
			d.SetPartial("config")
		})
	}

	if d.HasChange("security_group_ids") {
		securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

		req.SecurityGroupIds = securityGroupIds
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "security_group_ids")

		onDone = append(onDone, func() {
			d.SetPartial("security_group_ids")
		})
	}

	if d.HasChange("maintenance_window") {
		mw, err := expandRedisMaintenanceWindow(d)
		if err != nil {
			return err
		}
		req.MaintenanceWindow = mw
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "maintenance_window")

		onDone = append(onDone, func() {
			d.SetPartial("maintenance_window")
		})
	}

	err := makeRedisClusterUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	for _, f := range onDone {
		f()
	}
	return nil
}

func updateRedisClusterHosts(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	sharded := d.Get("sharded").(bool)

	currHosts, err := listRedisHosts(ctx, config, d)
	if err != nil {
		return err
	}

	targetHosts, err := expandRedisHosts(d)
	if err != nil {
		return fmt.Errorf("Error while expanding hosts on Redis Cluster create: %s", err)
	}

	currShards, err := listRedisShards(ctx, config, d)
	if err != nil {
		return err
	}

	toDelete, toAdd := redisHostsDiff(currHosts, targetHosts)

	ctx, cancel = context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	for shardName, specs := range toAdd {
		shardExists := false
		for _, s := range currShards {
			if s.Name == shardName {
				shardExists = true
			}
		}
		if sharded && !shardExists {
			err = createRedisShard(ctx, config, d, shardName, specs)
			if err != nil {
				return err
			}
		} else {
			err = createRedisHosts(ctx, config, d, specs)
			if err != nil {
				return err
			}
		}
	}

	for shardName, fqdns := range toDelete {
		deleteShard := true
		for _, th := range targetHosts {
			if th.ShardName == shardName {
				deleteShard = false
			}
		}
		if sharded && deleteShard {
			err = deleteRedisShard(ctx, config, d, shardName)
			if err != nil {
				return err
			}
		} else {
			err = deleteRedisHosts(ctx, config, d, fqdns)
			if err != nil {
				return err
			}
		}
	}

	d.SetPartial("host")
	return nil
}

func updateRedisMaintenanceWindow(ctx context.Context, config *Config, d *schema.ResourceData, mw *redis.MaintenanceWindow) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Redis().Cluster().Update(ctx, &redis.UpdateClusterRequest{
			ClusterId:         d.Id(),
			MaintenanceWindow: mw,
			UpdateMask:        &field_mask.FieldMask{Paths: []string{"maintenance_window"}},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update maintenance window in Redis Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating maintenance window in Redis Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func listRedisHosts(ctx context.Context, config *Config, d *schema.ResourceData) ([]*redis.Host, error) {
	hosts := []*redis.Host{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Redis().Cluster().ListHosts(ctx, &redis.ListClusterHostsRequest{
			ClusterId: d.Id(),
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for '%s': %s", d.Id(), err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return hosts, nil
}

func listRedisShards(ctx context.Context, config *Config, d *schema.ResourceData) ([]*redis.Shard, error) {
	shards := []*redis.Shard{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Redis().Cluster().ListShards(ctx, &redis.ListClusterShardsRequest{
			ClusterId: d.Id(),
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of shards for '%s': %s", d.Id(), err)
		}
		shards = append(shards, resp.Shards...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return shards, nil
}

func createRedisShard(ctx context.Context, config *Config, d *schema.ResourceData, shardName string, hostSpecs []*redis.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Redis().Cluster().AddShard(ctx, &redis.AddClusterShardRequest{
			ClusterId: d.Id(),
			ShardName: shardName,
			HostSpecs: hostSpecs,
		}),
	)
	if err != nil {
		return fmt.Errorf("Error while requesting API to add shard to Redis Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while adding shard to Redis Cluster %q: %s", d.Id(), err)
	}
	op, err = config.sdk.WrapOperation(
		config.sdk.MDB().Redis().Cluster().Rebalance(ctx, &redis.RebalanceClusterRequest{
			ClusterId: d.Id(),
		}),
	)
	if err != nil {
		return fmt.Errorf("Error while requesting API to rebalance the Redis Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while rebalancing the Redis Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createRedisHosts(ctx context.Context, config *Config, d *schema.ResourceData, specs []*redis.HostSpec) error {
	for _, hs := range specs {
		op, err := config.sdk.WrapOperation(
			config.sdk.MDB().Redis().Cluster().AddHosts(ctx, &redis.AddClusterHostsRequest{
				ClusterId: d.Id(),
				HostSpecs: []*redis.HostSpec{hs},
			}),
		)
		if err != nil {
			return fmt.Errorf("Error while requesting API to add host to Redis Cluster %q: %s", d.Id(), err)
		}
		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error while adding host to Redis Cluster %q: %s", d.Id(), err)
		}
	}
	return nil
}

func deleteRedisShard(ctx context.Context, config *Config, d *schema.ResourceData, shardName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Redis().Cluster().DeleteShard(ctx, &redis.DeleteClusterShardRequest{
			ClusterId: d.Id(),
			ShardName: shardName,
		}),
	)
	if err != nil {
		return fmt.Errorf("Error while requesting API to delete shard from Redis Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while deleting shard from Redis Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteRedisHosts(ctx context.Context, config *Config, d *schema.ResourceData, fqdns []string) error {
	for _, fqdn := range fqdns {
		op, err := config.sdk.WrapOperation(
			config.sdk.MDB().Redis().Cluster().DeleteHosts(ctx, &redis.DeleteClusterHostsRequest{
				ClusterId: d.Id(),
				HostNames: []string{fqdn},
			}),
		)
		if err != nil {
			return fmt.Errorf("Error while requesting API to delete host %s from Redis Cluster %q: %s", fqdn, d.Id(), err)
		}
		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("Error while deleting host %s from Redis Cluster %q: %s", fqdn, d.Id(), err)
		}
	}
	return nil
}

func makeRedisClusterUpdateRequest(req *redis.UpdateClusterRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Redis().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Redis Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Redis Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func resourceYandexMDBRedisClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Redis Cluster %q", d.Id())

	req := &redis.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Redis().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Redis Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Redis Cluster %q", d.Id())
	return nil
}
