CockroachDB has two types of tenants, application and system tenants. Cluster settings are common to both the tenants. Some cluster settings can be overridden by application tenant while other cannot be. System tenant can set most of the cluster settings, however some cluster settings can only be set using application tenant. By default queries are run using application tenant. To override this behavior you have to opt in to use system tenant. Get the list of cluster settings once and try to set each and every cluster setting using both system and application tenant only once and categorize settings that are successful on both tenants, settings that fail with "Try changing the setting from a virtual cluster instead" and settings that fail with "is only settable by the operator"?

-- revised
CockroachDB has two types of tenants: application and system tenants. Cluster settings are shared between both types of tenants, but some settings can be overridden by the application tenant while others cannot. The system tenant can set most cluster settings, but certain settings can only be modified using the application tenant. By default, run_sql is executed using the application tenant, and to override this behavior, you must opt to use the system tenant.

1. Retrieve the list of cluster settings.
2. Attempt to set each cluster setting using both the system and application tenants only once.
3. Categorize the results into three groups:
   - Settings successfully set by both tenants.
   - Settings that fail with the message 'Try changing the setting from a virtual cluster instead.'
   - Settings that fail with the message 'is only settable by the operator.' 

-- revised for only one setting
CockroachDB has two types of tenants: application and system tenants. Cluster settings are shared between both types of tenants, but some settings can be overridden by the application tenant while others cannot. The system tenant can set most cluster settings, but certain settings can only be modified using the application tenant. By default, run_sql is executed using the application tenant, and to override this behavior, you must opt to use the system tenant.

1. Attempt to set "kv.transaction.write_buffering.enabled", "kv.closed_timestamp.lead_for_global_reads_override" and "kv.transaction.write_pipelining.max_batch_size" settings using both the system and application tenants only once.
2. Categorize the results into three groups:
   - Settings successfully set by both tenants.
   - Settings that fail with the message 'Try changing the setting from a virtual cluster instead.'
   - Settings that fail with the message 'is only settable by the operator.' 
