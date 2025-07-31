#!/usr/bin/env bash

# 使用给定的名字 初始化一个新的微服务

new_svc=$1
if [ -z "$new_svc" ]; then
  echo "Usage: new_svc.sh <new_svc>"
  exit 1
fi

to_std_svc_name(){
  local my_var=$1
  # 去掉所有下划线并将字符串转换为小写
  cleaned_var=$(echo "$my_var" | tr -d '_' | tr 'A-Z' 'a-z')
  echo "$cleaned_var"
}

replace_in_files() {
  local dir=$1
  local pattern=$2
  local replacement=$3

  # Find all files and apply sed only to text files
#  echo "Replacing $pattern with $replacement in $dir"
  find "$dir" -type f -exec sed -i '' "s@$pattern@$replacement@g" {} +
}

new_svc=$(to_std_svc_name "$new_svc")
capitalized_svc="$(echo "$new_svc" | awk '{print toupper(substr($0, 1, 1)) substr($0, 2)}')"


cp -r service/template service/$new_svc
newsvc_pb_dir_name="$new_svc"pb
mkdir -p proto/svc/$newsvc_pb_dir_name && cp -r proto/svc/templatepb/* proto/svc/$newsvc_pb_dir_name

replace_in_files "service/$new_svc" "Template" "$capitalized_svc"
replace_in_files "proto/svc/$newsvc_pb_dir_name" "Template" "$capitalized_svc"

replace_in_files "service/$new_svc" "template" "$new_svc"
replace_in_files "proto/svc/$newsvc_pb_dir_name" "template" "$new_svc"

git add service/$new_svc 2>/dev/null
git add proto/svc/$newsvc_pb_dir_name 2>/dev/null