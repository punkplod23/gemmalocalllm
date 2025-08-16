import torch
from unsloth import FastLanguageModel
from datasets import load_dataset
from trl import SFTTrainer
from transformers import TrainingArguments

print("Unsloth training script started.")

# 1. Load a model
max_seq_length = 2048
dtype = None # None for auto detection. Float16 for Tesla T4, V100; Bfloat16 for Ampere+
load_in_4bit = True # Use 4bit quantization to reduce memory usage.

print("Loading base model...")
model, tokenizer = FastLanguageModel.from_pretrained(
    model_name = "unsloth/gemma-2b-it-bnb-4bit", # Using a pre-quantized model from Unsloth's hub
    max_seq_length = max_seq_length,
    dtype = dtype,
    load_in_4bit = load_in_4bit,
)
print("✅ Model loaded.")

# 2. Add LoRA adapters for fine-tuning
model = FastLanguageModel.get_peft_model(
    model,
    r = 16, # Choose any number > 0
    target_modules = ["q_proj", "k_proj", "v_proj", "o_proj",
                      "gate_proj", "up_proj", "down_proj",],
    lora_alpha = 16,
    lora_dropout = 0,
    bias = "none",
    use_gradient_checkpointing = True,
    random_state = 3407,
)
print("✅ LoRA adapters configured.")

# 3. Load a dataset
# In a real scenario, you would load your data from /app/data
# For example: dataset = load_dataset("csv", data_files={"train": "/app/data/my_training_data.csv"})
alpaca_prompt = """Below is an instruction that describes a task. Write a response that appropriately completes the request.

### Instruction:
{}

### Response:
{}"""

def formatting_prompts_func(examples):
    instructions = examples["instruction"]
    outputs      = examples["output"]
    texts = [alpaca_prompt.format(inst, out) for inst, out in zip(instructions, outputs)]
    return { "text" : texts, }

print("Loading and formatting dataset...")
dataset = load_dataset("yahma/alpaca-cleaned", split = "train")
dataset = dataset.map(formatting_prompts_func, batched = True,)
print("✅ Dataset ready.")

# 4. Configure and run the trainer
trainer = SFTTrainer(
    model = model,
    tokenizer = tokenizer,
    train_dataset = dataset,
    dataset_text_field = "text",
    max_seq_length = max_seq_length,
    dataset_num_proc = 2,
    packing = False, # Can make training 5x faster for short sequences.
    args = TrainingArguments(
        per_device_train_batch_size = 2,
        gradient_accumulation_steps = 4,
        warmup_steps = 5,
        max_steps = 60, # Set to a higher number for a real training run
        learning_rate = 2e-4,
        fp16 = not torch.cuda.is_bf16_supported(),
        bf16 = torch.cuda.is_bf16_supported(),
        logging_steps = 1,
        optim = "adamw_8bit",
        output_dir = "outputs",
    ),
)

print("🚀 Starting training...")
trainer.train()
print("✅ Training complete!")

print("💾 Saving LoRA adapter to /app/lora_adapters/gemma-2b-it-lora")
model.save_pretrained("/app/lora_adapters/gemma-2b-it-lora")
print("✅ Adapter saved successfully!")