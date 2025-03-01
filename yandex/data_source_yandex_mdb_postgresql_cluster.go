package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBPostgreSQLCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBPostgreSQLClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"config": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"access": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_lens": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"web_sql": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"autofailover": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"backup_window_start": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"minutes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"performance_diagnostics": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"sessions_sampling_interval": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"statements_sampling_interval": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"pooler_config": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pool_discard": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"pooling_mode": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"disk_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"resource_preset_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"postgresql_config": {
							Type:             schema.TypeMap,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGSettingsFieldsInfo),
							ValidateFunc:     generateMapSchemaValidateFunc(mdbPGSettingsFieldsInfo),
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"extension": {
							Type:     schema.TypeSet,
							Set:      pgExtensionHash,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"version": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"lc_collate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lc_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner": {
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
			"environment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host": {
				Type:     schema.TypeList,
				MinItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"replication_source": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grants": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"login": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password": {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
						"permission": {
							Type:     schema.TypeSet,
							Computed: true,
							Set:      pgUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"conn_limit": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"settings": {
							Type:             schema.TypeMap,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGUserSettingsFieldsInfo),
							ValidateFunc:     generateMapSchemaValidateFunc(mdbPGUserSettingsFieldsInfo),
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Computed: true,
			},
			"maintenance_window": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"day": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hour": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexMDBPostgreSQLClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.PostgreSQLClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source PostgreSQL Cluster by name: %v", err)
		}
	}

	cluster, err := config.sdk.MDB().PostgreSQL().Cluster().Get(ctx, &postgresql.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", clusterID))
	}

	pgClusterConfig, err := flattenPGClusterConfig(cluster.Config, d)
	if err != nil {
		return err
	}
	if err := d.Set("config", pgClusterConfig); err != nil {
		return err
	}

	hosts, err := listPGHosts(ctx, config, clusterID)
	if err != nil {
		return err
	}
	hs, _, err := flattenPGHosts(d, hosts, true)
	if err != nil {
		return err
	}
	if err := d.Set("host", hs); err != nil {
		return err
	}

	databases, err := listPGDatabases(ctx, config, clusterID)
	if err != nil {
		return err
	}
	dbs := flattenPGDatabases(databases)
	if err := d.Set("database", dbs); err != nil {
		return err
	}

	users, err := listPGUsers(ctx, config, clusterID)
	if err != nil {
		return err
	}
	us, err := flattenPGUsers(users, nil, mdbPGUserSettingsFieldsInfo)
	if err != nil {
		return err
	}
	if err := d.Set("user", us); err != nil {
		return err
	}

	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	maintenanceWindow, err := flattenPGMaintenanceWindow(cluster.MaintenanceWindow)
	if err != nil {
		return err
	}

	if err := d.Set("maintenance_window", maintenanceWindow); err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("cluster_id", cluster.Id)
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)

	d.SetId(cluster.Id)
	return nil
}
