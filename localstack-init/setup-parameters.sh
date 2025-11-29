#!/bin/bash

# 環境変数からPARAM_で始まるものを取得し、parameter-storeにセットするスクリプト
echo "Setting up Parameter Store parameters..."

# PARAMプレフィックスの環境変数をparameter形式(github-app-)に変換
convert_env_to_param() {
  local env_name
  local param_name
  local full_param_name
  local param_type
  local param_value

  env_name="$1"
  param_value="$2"

  # 環境変数をケバブ形式に変換(ex. PARAM_PREFIX -> param-prefix)
  param_name=$(echo "$env_name" | tr '[:upper:]' '[:lower:]' | tr '_' '-')

  full_param_name="${SSM_PARAM_PREFIX}/$param_name"

  # 機密情報の場合は暗号化する
  param_type="String"
  if [[ "$param_name" == *"key"* ]] || [[ "$param_name" == *"secret"* ]]; then
    param_type="SecureString"
  fi

  # 空でなければパラメータをセットする
  if [[ -n "$param_value" ]]; then
    echo "Creating parameter: $full_param_name"
    awslocal ssm put-parameter \
      --name "$full_param_name" \
      --value "$param_value" \
      --type "$param_type" \
      --description "GitHub App $param_name"
  else
    echo "Skipping $full_param_name (empty value)"
  fi
}

# メインの処理部分
env_name=""
env_value=""

for env_name in $(env | grep '^PARAM_' | cut -d= -f1); do
  env_value="${!env_name}"
  convert_env_to_param "$env_name" "$env_value"
done

echo "Parameter Store setup complete!"

# List all parameters to verify
awslocal ssm describe-parameters --query 'Parameters[*].[Name,Type]' --output table