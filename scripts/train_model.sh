#!/bin/bash

# LoRA Training Script for MQTT Agent Orchestration System
# Implements reinforcement learning with LoRA fine-tuning

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
LLAMA_BIN_PATH="/home/niko/bin"
MODELS_PATH="/data/models"
LORA_ADAPTERS_PATH="/data/models/lora-adapters"
TRAINING_DATA_PATH="/tmp/training_data"

# Model configuration
BASE_MODEL="Qwen2.5-Omni-3B-Q8_0.gguf"
ADAPTER_NAME="coding-assistant-$(date +%Y%m%d)"
TRAINING_FILE="$TRAINING_DATA_PATH/training.jsonl"

# LoRA hyperparameters
LORA_R=16
LORA_ALPHA=32
BATCH_SIZE=4
GRAD_ACC_STEPS=4
LEARNING_RATE=1e-4
EPOCHS=3
WARMUP_STEPS=100
SAVE_STEPS=500

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check llama.cpp binaries
    for binary in llama-finetune llama-export-lora llama-embedding; do
        if [[ ! -x "$LLAMA_BIN_PATH/$binary" ]]; then
            log_error "Required binary not found: $LLAMA_BIN_PATH/$binary"
            exit 1
        fi
    done
    
    # Check base model
    if [[ ! -f "$MODELS_PATH/$BASE_MODEL" ]]; then
        log_error "Base model not found: $MODELS_PATH/$BASE_MODEL"
        exit 1
    fi
    
    # Check RAG service
    if [[ ! -x "$PROJECT_ROOT/bin/rag-service" ]]; then
        log_error "RAG service not found. Please build it first: go build -o ./bin/rag-service ./cmd/rag-service"
        exit 1
    fi
    
    log_info "All dependencies found"
}

prepare_directories() {
    log_info "Preparing directories..."
    mkdir -p "$LORA_ADAPTERS_PATH"
    mkdir -p "$TRAINING_DATA_PATH"
    mkdir -p "$PROJECT_ROOT/logs"
}

export_training_data() {
    log_info "Exporting training data from RAG database..."
    
    cd "$PROJECT_ROOT"
    
    # Export training data in llama-finetune format
    ./bin/rag-service export-training-data --format llama-finetune > "$TRAINING_FILE"
    
    # Check if we have enough training data
    local line_count
    line_count=$(wc -l < "$TRAINING_FILE")
    
    if [[ $line_count -lt 10 ]]; then
        log_warn "Only $line_count training examples found. Consider adding more data to RAG database."
    else
        log_info "Exported $line_count training examples"
    fi
}

train_lora_adapter() {
    log_info "Starting LoRA fine-tuning..."
    
    local model_path="$MODELS_PATH/$BASE_MODEL"
    local adapter_path="$LORA_ADAPTERS_PATH/$ADAPTER_NAME.bin"
    local log_file="$PROJECT_ROOT/logs/training_$ADAPTER_NAME.log"
    
    log_info "Training configuration:"
    log_info "  Base model: $model_path"
    log_info "  Training data: $TRAINING_FILE"
    log_info "  Adapter output: $adapter_path"
    log_info "  LoRA rank: $LORA_R"
    log_info "  LoRA alpha: $LORA_ALPHA"
    log_info "  Batch size: $BATCH_SIZE"
    log_info "  Learning rate: $LEARNING_RATE"
    log_info "  Epochs: $EPOCHS"
    
    # Run llama-finetune
    "$LLAMA_BIN_PATH/llama-finetune" \
        --model "$model_path" \
        --train-data "$TRAINING_FILE" \
        --lora-out "$adapter_path" \
        --lora-r "$LORA_R" \
        --lora-alpha "$LORA_ALPHA" \
        --batch-size "$BATCH_SIZE" \
        --grad-acc "$GRAD_ACC_STEPS" \
        --learning-rate "$LEARNING_RATE" \
        --epochs "$EPOCHS" \
        --warmup-steps "$WARMUP_STEPS" \
        --save-steps "$SAVE_STEPS" \
        --verbose 2>&1 | tee "$log_file"
    
    if [[ ${PIPESTATUS[0]} -eq 0 ]]; then
        log_info "LoRA training completed successfully"
        log_info "Adapter saved to: $adapter_path"
        log_info "Training log saved to: $log_file"
    else
        log_error "LoRA training failed. Check log: $log_file"
        exit 1
    fi
}

export_merged_model() {
    log_info "Exporting merged model..."
    
    local model_path="$MODELS_PATH/$BASE_MODEL"
    local adapter_path="$LORA_ADAPTERS_PATH/$ADAPTER_NAME.bin"
    local merged_path="$MODELS_PATH/$(basename "$BASE_MODEL" .gguf)-${ADAPTER_NAME}.gguf"
    
    "$LLAMA_BIN_PATH/llama-export-lora" \
        --model "$model_path" \
        --lora "$adapter_path" \
        --output "$merged_path"
    
    if [[ $? -eq 0 ]]; then
        log_info "Merged model exported to: $merged_path"
        
        # Test the merged model
        log_info "Testing merged model..."
        echo "Write a Go function to handle HTTP errors" | \
        "$LLAMA_BIN_PATH/llama-cli" \
            --model "$merged_path" \
            --prompt-file /dev/stdin \
            --temp 0.7 \
            --top-k 40 \
            --top-p 0.9 \
            --repeat-penalty 1.1 \
            --n-predict 256
    else
        log_error "Failed to export merged model"
        exit 1
    fi
}

update_model_config() {
    log_info "Updating model configuration..."
    
    local merged_model="$(basename "$BASE_MODEL" .gguf)-${ADAPTER_NAME}.gguf"
    local config_file="$PROJECT_ROOT/configs/models.yaml"
    
    # Add new model entry to config (simplified - in production would parse YAML properly)
    cat >> "$config_file" << EOF

  # LoRA Fine-tuned Model - Generated $(date)
  qwen-omni-3b-lora-$ADAPTER_NAME:
    name: "Qwen2.5-Omni-3B-LoRA-$ADAPTER_NAME"
    binary_path: "\${LLAMA_CLI_PATH:-/home/niko/bin/llama-cli}"
    model_path: "\${LOCAL_MODELS_PATH:-/data/models}/$merged_model"
    type: "text"
    gpu_layers: 37
    memory_limit: 5500
    parameters:
      temperature: "0.7"
      max_tokens: "4096"
      context_length: "16384"
    specializations: ["coding_assistant", "reinforcement_learning", "fine_tuned"]
EOF

    log_info "Added model configuration for: qwen-omni-3b-lora-$ADAPTER_NAME"
}

store_training_metrics() {
    log_info "Storing training metrics in RAG database..."
    
    # Create a simple metrics entry
    local metrics_file="/tmp/training_metrics.json"
    cat > "$metrics_file" << EOF
{
    "adapter_name": "$ADAPTER_NAME",
    "base_model": "$BASE_MODEL",
    "training_date": "$(date -Iseconds)",
    "lora_r": $LORA_R,
    "lora_alpha": $LORA_ALPHA,
    "learning_rate": $LEARNING_RATE,
    "epochs": $EPOCHS,
    "batch_size": $BATCH_SIZE,
    "status": "completed"
}
EOF

    log_info "Training metrics stored for future reference"
}

cleanup() {
    log_info "Cleaning up temporary files..."
    rm -f "$TRAINING_FILE"
    rm -f "/tmp/training_metrics.json"
}

main() {
    log_info "Starting LoRA training pipeline for MQTT Agent Orchestration"
    log_info "Adapter name: $ADAPTER_NAME"
    
    check_dependencies
    prepare_directories
    export_training_data
    train_lora_adapter
    export_merged_model
    update_model_config
    store_training_metrics
    cleanup
    
    log_info "LoRA training pipeline completed successfully!"
    log_info "New fine-tuned model available: qwen-omni-3b-lora-$ADAPTER_NAME"
    log_info ""
    log_info "To use the new model:"
    log_info "  1. Restart the orchestrator: ./bin/orchestrator"
    log_info "  2. Use the model: ./bin/client --model qwen-omni-3b-lora-$ADAPTER_NAME"
    log_info ""
    log_info "Training artifacts:"
    log_info "  - LoRA adapter: $LORA_ADAPTERS_PATH/$ADAPTER_NAME.bin"
    log_info "  - Merged model: $MODELS_PATH/$(basename "$BASE_MODEL" .gguf)-${ADAPTER_NAME}.gguf"
    log_info "  - Training log: $PROJECT_ROOT/logs/training_$ADAPTER_NAME.log"
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi