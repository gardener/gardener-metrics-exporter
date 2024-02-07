package metrics

const (
	metricGardenProjectsStatus = "garden_projects_status"
	metricGardenUsersSum       = "garden_users_total"

	// Seed metric
	metricGardenManagedSeedInfo = "garden_managed_seed_info"
	metricGardenSeedInfo        = "garden_seed_info"
	metricGardenSeedCondition   = "garden_seed_condition"
	metricGardenSeedCapacity    = "garden_seed_capacity"
	metricGardenSeedUsage       = "garden_seed_usage"

	// Shoot metric (available also for Shoots which act as Seed).
	metricGardenShootCondition                = "garden_shoot_condition"
	metricGardenShootCreation                 = "garden_shoot_creation_timestamp"
	metricGardenShootHibernated               = "garden_shoot_hibernated"
	metricGardenShootInfo                     = "garden_shoot_info"
	metricGardenShootNodeMaxTotal             = "garden_shoot_node_max_total"
	metricGardenShootNodeMinTotal             = "garden_shoot_node_min_total"
	metricGardenShootOperationProgressPercent = "garden_shoot_operation_progress_percent"
	metricGardenShootOperationState           = "garden_shoot_operation_states"
	metricGardenShootWorkerNodeMaxTotal       = "garden_shoot_worker_node_max_total"
	metricGardenShootWorkerNodeMinTotal       = "garden_shoot_worker_node_min_total"

	// Aggregated Shoot metrics (exclude Shoots which act as Seed).
	metricGardenOperationsTotal = "garden_shoot_operations_total"
	metricGardenShootNodeInfo   = "garden_shoot_node_info"
)
