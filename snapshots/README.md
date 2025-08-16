# System Snapshots (`snapshots/`)

## Overview

The `snapshots/` directory contains system snapshots and backup data for the MQTT Agent Orchestration System. This directory implements the **"Fail Safe"** principle by providing comprehensive backup and recovery capabilities to ensure system resilience and data protection.

## Architecture Philosophy

Following our **"Excellence through Rigor"** philosophy, the snapshot system is:
- **Comprehensive**: Complete system state capture
- **Incremental**: Efficient storage with incremental backups
- **Verifiable**: Checksum validation for data integrity
- **Recoverable**: Fast and reliable restoration procedures

## Snapshot Types

### 1. System State Snapshots

**Purpose**: Complete system state capture for disaster recovery.

**Components Captured**:
- Configuration files
- Database state (Qdrant collections)
- Model states and metadata
- Worker status and health data
- System logs and metrics

**Snapshot Structure**:
```
snapshots/
├── system/
│   ├── 2024-01-15_10-30-00/
│   │   ├── configs/
│   │   │   ├── models.yaml
│   │   │   ├── mcp.yaml
│   │   │   └── system.yaml
│   │   ├── storage/
│   │   │   ├── collections/
│   │   │   ├── aliases/
│   │   │   └── raft_state.json
│   │   ├── models/
│   │   │   ├── metadata.json
│   │   │   └── state.json
│   │   ├── workers/
│   │   │   ├── status.json
│   │   │   └── health.json
│   │   ├── logs/
│   │   │   ├── orchestrator.log
│   │   │   ├── role-worker.log
│   │   │   └── system.log
│   │   └── metadata.json
│   └── latest -> 2024-01-15_10-30-00/
```

**Metadata Structure**:
```json
{
  "snapshot_id": "2024-01-15_10-30-00",
  "timestamp": "2024-01-15T10:30:00Z",
  "type": "system",
  "version": "1.0.0",
  "components": {
    "orchestrator": "running",
    "workers": 4,
    "collections": 14,
    "models": 5
  },
  "checksums": {
    "configs": "sha256:abc123...",
    "storage": "sha256:def456...",
    "models": "sha256:ghi789..."
  },
  "size_bytes": 1073741824,
  "compression_ratio": 0.75
}
```

### 2. Configuration Snapshots

**Purpose**: Configuration state capture for version control and rollback.

**Components Captured**:
- All configuration files
- Environment variables
- System settings
- User preferences

**Configuration Snapshot Structure**:
```
snapshots/
├── configs/
│   ├── 2024-01-15_10-30-00/
│   │   ├── models.yaml
│   │   ├── mcp.yaml
│   │   ├── system.yaml
│   │   ├── environment.json
│   │   └── metadata.json
│   └── latest -> 2024-01-15_10-30-00/
```

### 3. Data Snapshots

**Purpose**: Database and storage state capture for data protection.

**Components Captured**:
- Qdrant vector collections
- RAG knowledge base
- System aliases and mappings
- User data and preferences

**Data Snapshot Structure**:
```
snapshots/
├── data/
│   ├── 2024-01-15_10-30-00/
│   │   ├── qdrant/
│   │   │   ├── collections/
│   │   │   ├── snapshots/
│   │   │   └── wal/
│   │   ├── storage/
│   │   │   ├── collections/
│   │   │   └── aliases/
│   │   └── metadata.json
│   └── latest -> 2024-01-15_10-30-00/
```

### 4. Model Snapshots

**Purpose**: Model state and metadata capture for model management.

**Components Captured**:
- Model metadata and configuration
- Model performance metrics
- Model loading states
- GPU memory allocation

**Model Snapshot Structure**:
```
snapshots/
├── models/
│   ├── 2024-01-15_10-30-00/
│   │   ├── metadata.json
│   │   ├── performance.json
│   │   ├── states.json
│   │   └── gpu_usage.json
│   └── latest -> 2024-01-15_10-30-00/
```

## Snapshot Management

### Creating Snapshots

**Manual Snapshot Creation**:
```bash
# Create system snapshot
./scripts/create_snapshot.sh --type system --name "pre-deployment"

# Create configuration snapshot
./scripts/create_snapshot.sh --type config --name "config-backup"

# Create data snapshot
./scripts/create_snapshot.sh --type data --name "data-backup"

# Create model snapshot
./scripts/create_snapshot.sh --type models --name "model-state"
```

**Automated Snapshot Creation**:
```bash
# Schedule regular snapshots
crontab -e

# Daily system snapshot at 2 AM
0 2 * * * /home/niko/Dev/mqtt_agent_orchestration/scripts/create_snapshot.sh --type system --auto

# Weekly full backup on Sunday at 3 AM
0 3 * * 0 /home/niko/Dev/mqtt_agent_orchestration/scripts/create_snapshot.sh --type full --auto
```

**Snapshot Script Configuration**:
```bash
#!/bin/bash
# scripts/create_snapshot.sh

# Snapshot configuration
SNAPSHOT_DIR="/home/niko/Dev/mqtt_agent_orchestration/snapshots"
SNAPSHOT_TYPES=("system" "config" "data" "models" "full")
COMPRESSION_LEVEL=9
RETENTION_DAYS=30

# Create snapshot function
create_snapshot() {
    local type="$1"
    local name="$2"
    local timestamp=$(date +%Y-%m-%d_%H-%M-%S)
    local snapshot_path="$SNAPSHOT_DIR/$type/$timestamp"
    
    echo "Creating $type snapshot: $name"
    
    # Create snapshot directory
    mkdir -p "$snapshot_path"
    
    # Capture system state based on type
    case "$type" in
        "system")
            capture_system_state "$snapshot_path"
            ;;
        "config")
            capture_config_state "$snapshot_path"
            ;;
        "data")
            capture_data_state "$snapshot_path"
            ;;
        "models")
            capture_model_state "$snapshot_path"
            ;;
        "full")
            capture_full_state "$snapshot_path"
            ;;
    esac
    
    # Create metadata
    create_snapshot_metadata "$snapshot_path" "$type" "$name"
    
    # Compress snapshot
    compress_snapshot "$snapshot_path"
    
    # Update latest symlink
    update_latest_symlink "$type" "$timestamp"
    
    echo "Snapshot created: $snapshot_path"
}
```

### Snapshot Verification

**Integrity Verification**:
```bash
# Verify snapshot integrity
./scripts/verify_snapshot.sh --snapshot 2024-01-15_10-30-00

# Verify all snapshots
./scripts/verify_snapshot.sh --all

# Verify specific components
./scripts/verify_snapshot.sh --snapshot 2024-01-15_10-30-00 --component configs
```

**Verification Script**:
```bash
#!/bin/bash
# scripts/verify_snapshot.sh

verify_snapshot() {
    local snapshot_path="$1"
    local metadata_file="$snapshot_path/metadata.json"
    
    echo "Verifying snapshot: $snapshot_path"
    
    # Check metadata exists
    if [[ ! -f "$metadata_file" ]]; then
        echo "ERROR: Metadata file not found"
        return 1
    fi
    
    # Verify checksums
    local checksums=$(jq -r '.checksums | to_entries[] | "\(.key):\(.value)"' "$metadata_file")
    
    for checksum in $checksums; do
        local component=$(echo "$checksum" | cut -d: -f1)
        local expected_hash=$(echo "$checksum" | cut -d: -f2)
        local component_path="$snapshot_path/$component"
        
        if [[ -d "$component_path" ]]; then
            local actual_hash=$(calculate_checksum "$component_path")
            if [[ "$actual_hash" != "$expected_hash" ]]; then
                echo "ERROR: Checksum mismatch for $component"
                return 1
            fi
        fi
    done
    
    echo "Snapshot verification passed"
    return 0
}
```

### Snapshot Restoration

**System Restoration**:
```bash
# Restore system from snapshot
./scripts/restore_snapshot.sh --snapshot 2024-01-15_10-30-00 --type system

# Restore specific components
./scripts/restore_snapshot.sh --snapshot 2024-01-15_10-30-00 --component configs

# Restore with verification
./scripts/restore_snapshot.sh --snapshot 2024-01-15_10-30-00 --verify
```

**Restoration Script**:
```bash
#!/bin/bash
# scripts/restore_snapshot.sh

restore_snapshot() {
    local snapshot_path="$1"
    local restore_type="$2"
    local verify="$3"
    
    echo "Restoring from snapshot: $snapshot_path"
    
    # Stop system services
    stop_system_services
    
    # Create backup of current state
    create_backup_before_restore
    
    # Restore based on type
    case "$restore_type" in
        "system")
            restore_system_state "$snapshot_path"
            ;;
        "config")
            restore_config_state "$snapshot_path"
            ;;
        "data")
            restore_data_state "$snapshot_path"
            ;;
        "models")
            restore_model_state "$snapshot_path"
            ;;
        "full")
            restore_full_state "$snapshot_path"
            ;;
    esac
    
    # Verify restoration if requested
    if [[ "$verify" == "true" ]]; then
        verify_restoration "$snapshot_path"
    fi
    
    # Start system services
    start_system_services
    
    echo "Restoration completed"
}
```

## Backup Strategies

### Incremental Backups

**Incremental Backup Strategy**:
```bash
# Full backup on Sunday
0 3 * * 0 /scripts/create_snapshot.sh --type full --name "weekly-full"

# Incremental backup daily
0 2 * * 1-6 /scripts/create_snapshot.sh --type incremental --name "daily-incremental"

# Differential backup on Wednesday
0 4 * * 3 /scripts/create_snapshot.sh --type differential --name "weekly-differential"
```

**Incremental Backup Implementation**:
```bash
create_incremental_backup() {
    local base_snapshot="$1"
    local incremental_path="$2"
    
    # Find last full backup
    local last_full=$(find_last_full_backup)
    
    # Create incremental backup
    rsync -av --link-dest="$last_full" \
        --exclude="*.tmp" \
        --exclude="*.log" \
        /home/niko/Dev/mqtt_agent_orchestration/ \
        "$incremental_path/"
    
    # Create incremental metadata
    create_incremental_metadata "$incremental_path" "$last_full"
}
```

### Differential Backups

**Differential Backup Strategy**:
```bash
create_differential_backup() {
    local base_snapshot="$1"
    local differential_path="$2"
    
    # Find last full backup
    local last_full=$(find_last_full_backup)
    
    # Create differential backup
    tar --create --file="$differential_path.tar" \
        --listed-incremental="$last_full/incremental.list" \
        --exclude="*.tmp" \
        --exclude="*.log" \
        /home/niko/Dev/mqtt_agent_orchestration/
    
    # Compress differential backup
    gzip "$differential_path.tar"
}
```

## Snapshot Retention

### Retention Policies

**Retention Configuration**:
```yaml
# configs/retention.yaml
retention:
  system_snapshots:
    daily: 7      # Keep daily snapshots for 7 days
    weekly: 4     # Keep weekly snapshots for 4 weeks
    monthly: 12   # Keep monthly snapshots for 12 months
    
  config_snapshots:
    daily: 30     # Keep daily snapshots for 30 days
    weekly: 12    # Keep weekly snapshots for 12 weeks
    
  data_snapshots:
    daily: 7      # Keep daily snapshots for 7 days
    weekly: 4     # Keep weekly snapshots for 4 weeks
    monthly: 12   # Keep monthly snapshots for 12 months
    
  model_snapshots:
    daily: 7      # Keep daily snapshots for 7 days
    weekly: 4     # Keep weekly snapshots for 4 weeks
```

**Retention Script**:
```bash
#!/bin/bash
# scripts/cleanup_snapshots.sh

cleanup_snapshots() {
    local retention_config="$1"
    
    echo "Cleaning up old snapshots"
    
    # Load retention configuration
    local daily_retention=$(yq e '.retention.system_snapshots.daily' "$retention_config")
    local weekly_retention=$(yq e '.retention.system_snapshots.weekly' "$retention_config")
    local monthly_retention=$(yq e '.retention.system_snapshots.monthly' "$retention_config")
    
    # Remove old daily snapshots
    find "$SNAPSHOT_DIR/system" -name "*_daily_*" -mtime +$daily_retention -delete
    
    # Remove old weekly snapshots
    find "$SNAPSHOT_DIR/system" -name "*_weekly_*" -mtime +$((weekly_retention * 7)) -delete
    
    # Remove old monthly snapshots
    find "$SNAPSHOT_DIR/system" -name "*_monthly_*" -mtime +$((monthly_retention * 30)) -delete
    
    echo "Cleanup completed"
}
```

## Disaster Recovery

### Recovery Procedures

**Full System Recovery**:
```bash
# Full system recovery procedure
./scripts/disaster_recovery.sh --mode full --snapshot 2024-01-15_10-30-00

# Step-by-step recovery
./scripts/disaster_recovery.sh --mode step-by-step --snapshot 2024-01-15_10-30-00

# Recovery with validation
./scripts/disaster_recovery.sh --mode full --snapshot 2024-01-15_10-30-00 --validate
```

**Recovery Script**:
```bash
#!/bin/bash
# scripts/disaster_recovery.sh

disaster_recovery() {
    local mode="$1"
    local snapshot="$2"
    local validate="$3"
    
    echo "Starting disaster recovery: $mode"
    
    case "$mode" in
        "full")
            perform_full_recovery "$snapshot"
            ;;
        "step-by-step")
            perform_step_by_step_recovery "$snapshot"
            ;;
        "minimal")
            perform_minimal_recovery "$snapshot"
            ;;
    esac
    
    if [[ "$validate" == "true" ]]; then
        validate_recovery "$snapshot"
    fi
    
    echo "Disaster recovery completed"
}

perform_full_recovery() {
    local snapshot="$1"
    
    # Stop all services
    stop_all_services
    
    # Restore system state
    restore_system_state "$snapshot"
    
    # Restore configuration
    restore_configuration "$snapshot"
    
    # Restore data
    restore_data "$snapshot"
    
    # Restore models
    restore_models "$snapshot"
    
    # Start services
    start_all_services
    
    # Verify system health
    verify_system_health
}
```

### Recovery Testing

**Recovery Testing**:
```bash
# Test recovery procedure
./scripts/test_recovery.sh --snapshot 2024-01-15_10-30-00 --environment test

# Test recovery in isolated environment
./scripts/test_recovery.sh --snapshot 2024-01-15_10-30-00 --isolated

# Test recovery performance
./scripts/test_recovery.sh --snapshot 2024-01-15_10-30-00 --benchmark
```

## Monitoring and Alerting

### Snapshot Monitoring

**Health Monitoring**:
```bash
# Monitor snapshot health
./scripts/monitor_snapshots.sh --check-integrity

# Monitor snapshot age
./scripts/monitor_snapshots.sh --check-age

# Monitor snapshot size
./scripts/monitor_snapshots.sh --check-size
```

**Alerting Configuration**:
```yaml
# configs/snapshot_alerts.yaml
alerts:
  snapshot_age:
    warning_threshold: "24h"
    critical_threshold: "48h"
    
  snapshot_size:
    warning_threshold: "10GB"
    critical_threshold: "20GB"
    
  snapshot_integrity:
    check_interval: "1h"
    failure_threshold: 3
```

### Performance Monitoring

**Snapshot Performance Metrics**:
```bash
# Monitor snapshot creation time
./scripts/monitor_snapshots.sh --metrics creation-time

# Monitor snapshot restoration time
./scripts/monitor_snapshots.sh --metrics restoration-time

# Monitor snapshot compression ratio
./scripts/monitor_snapshots.sh --metrics compression-ratio
```

## Security Considerations

### Snapshot Security

**Access Control**:
```bash
# Set secure permissions
chmod 750 snapshots/
chmod 640 snapshots/*/metadata.json

# Restrict access to sensitive snapshots
chmod 600 snapshots/system/*/configs/
chmod 600 snapshots/system/*/storage/
```

**Encryption**:
```bash
# Encrypt sensitive snapshots
gpg --encrypt --recipient admin@example.com snapshots/system/2024-01-15_10-30-00.tar.gz

# Decrypt for restoration
gpg --decrypt snapshots/system/2024-01-15_10-30-00.tar.gz.gpg > snapshots/system/2024-01-15_10-30-00.tar.gz
```

### Data Protection

**PII Handling**:
```bash
# Sanitize snapshots for PII
./scripts/sanitize_snapshot.sh --snapshot 2024-01-15_10-30-00 --remove-pii

# Anonymize sensitive data
./scripts/sanitize_snapshot.sh --snapshot 2024-01-15_10-30-00 --anonymize
```

## Troubleshooting

### Common Issues

1. **Snapshot Creation Fails**: Check disk space and permissions
2. **Restoration Fails**: Verify snapshot integrity and dependencies
3. **Performance Issues**: Optimize compression and storage
4. **Corruption Issues**: Use checksum verification and recovery

### Debug Commands

```bash
# Debug snapshot creation
./scripts/create_snapshot.sh --debug --type system

# Debug snapshot restoration
./scripts/restore_snapshot.sh --debug --snapshot 2024-01-15_10-30-00

# Analyze snapshot contents
./scripts/analyze_snapshot.sh --snapshot 2024-01-15_10-30-00

# Check snapshot dependencies
./scripts/check_dependencies.sh --snapshot 2024-01-15_10-30-00
```

### Recovery Procedures

```bash
# Recover corrupted snapshot
./scripts/recover_snapshot.sh --snapshot 2024-01-15_10-30-00

# Rebuild snapshot metadata
./scripts/rebuild_metadata.sh --snapshot 2024-01-15_10-30-00

# Restore from backup
./scripts/restore_from_backup.sh --backup backup-20240115.tar.gz
```

## Future Enhancements

### Planned Features

- **Real-time Snapshots**: Continuous snapshot creation
- **Cloud Integration**: Cloud-based snapshot storage
- **Advanced Compression**: Improved compression algorithms
- **Snapshot Analytics**: Advanced snapshot analysis

### Extension Points

- **Custom Snapshot Types**: User-defined snapshot types
- **External Storage**: Integration with external storage systems
- **Snapshot Replication**: Multi-site snapshot replication
- **Snapshot Versioning**: Version control for snapshots

---

**Production Ready**: The snapshot system is designed for production use with comprehensive backup, verification, and recovery capabilities. It provides a robust foundation for system resilience and data protection.
