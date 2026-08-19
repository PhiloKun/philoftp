#!/usr/bin/env bash
#
# PhiloFTP 发布自动化脚本
# 流程：递增版本号 -> 构建全部平台产物 -> 创建 tag 并推双远端 -> 发布 GitHub + Gitee release 并上传附件
#
# 用法:
#   ./release.sh [patch|minor|major]   默认 patch
#   ./release.sh 1.2.3                 直接使用指定版本号
#   ./release.sh -n 1.2.3              仅递增/设定版本号,不构建与发布(干跑版本号)
#
# 前置:
#   - make, go, zip, makensis, python3 (Windows 安装包)
#   - git remote: origin(GitHub) / gitee(Gitee)
#   - GitHub CLI: gh (已登录)
#   - Gitee token: macOS Keychain (security find-internet-password -a PhiloKun -s gitee.com -w)
#
set -euo pipefail

# ---- 路径与配置 ----
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"
BINARY="philoftp"
VERSION_FILE="version.txt"
GH_REPO="PhiloKun/philoftp"
GITEE_REPO="PhiloKun/philoftp"
GITEE_API="https://gitee.com/api/v5/repos/${GITEE_REPO}"

DRYRUN=0
BUMP="${1:-patch}"
if [ "$BUMP" = "-n" ]; then
  DRYRUN=1
  BUMP="${2:-patch}"
fi

# ---- 读取当前版本 ----
read_version() {
  cat "$VERSION_FILE" 2>/dev/null | tr -d '[:space:]'
}

# ---- 版本递增 ----
bump_version() {
  local cur="$1"; local mode="$2"
  local major minor patch
  IFS='.' read -r major minor patch <<< "$cur"
  major="${major:-0}"; minor="${minor:-0}"; patch="${patch:-0}"
  case "$mode" in
    major) major=$((major+1)); minor=0; patch=0 ;;
    minor) minor=$((minor+1)); patch=0 ;;
    patch) patch=$((patch+1)) ;;
    *[0-9]*.[0-9]*.[0-9]*) echo "$mode"; return ;;  # 直接指定版本号
    *) echo "未知递增模式: $mode" >&2; exit 1 ;;
  esac
  echo "${major}.${minor}.${patch}"
}

# ---- Gitee token ----
gitee_token() {
  security find-internet-password -a PhiloKun -s gitee.com -w 2>/dev/null || true
}

# ---- 从 CHANGELOG.md 提取指定版本条目的发布说明 ----
# 结构约定：条目以 "## [X.Y.Z] - 日期" 开头，到下一个 "## " 或 "---" 前结束。
# 若 CHANGELOG 缺失或未找到对应版本，回退为通用说明。
changelog_notes() {
  local ver="$1"
  local notes=""
  if [ -f CHANGELOG.md ]; then
    notes="$(awk -v v="[$ver] " '
      $0 ~ "^## " { if (insec) exit; }
      $0 ~ "^---" { if (insec) exit; }
      index($0, "## " v) == 1 { insec=1; next }
      insec { print }
    ' CHANGELOG.md)"
  fi
  if [ -z "$(echo "$notes" | tr -d '[:space:]')" ]; then
    notes="PhiloFTP v${ver} 发布。单二进制内网 FTP 服务器 + Web 管理端。

> 提示：本版本未在 CHANGELOG.md 中找到对应条目，请补充发布说明后重新发布，或查看 Git 提交历史了解变更。"
  fi
  echo "$notes"
}

# 生成完整 release 说明：标题行 + CHANGELOG 条目
build_release_notes() {
  local ver="$1"
  local notes
  notes="$(changelog_notes "$ver")"
  printf '# PhiloFTP v%s\n\n%s\n' "$ver" "$notes"
}

echo "================================================"
echo " PhiloFTP 发布脚本"
echo "================================================"

CURRENT="$(read_version)"
echo "当前版本: $CURRENT"

if [ "$DRYRUN" -eq 1 ]; then
  NEW="$(bump_version "$CURRENT" "$BUMP")"
  echo "干跑: 新版本号将为 $NEW (写入 $VERSION_FILE 后退出)"
  echo "$NEW" > "$VERSION_FILE"
  exit 0
fi

NEW="$(bump_version "$CURRENT" "$BUMP")"
echo "新版本: $NEW"
echo "$NEW" > "$VERSION_FILE"
echo "已写入 $VERSION_FILE: $NEW"

# ---- 1. 同步文档 (README 等) 已在代码改动阶段完成, 此处仅确认 git 状态 ----
echo
echo "[1/5] 检查 git 工作区..."
if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "⚠ 存在未提交的改动, 自动提交文档/版本变更..."
  git add -A
  git commit -m "chore: 发布 v${NEW}" || true
fi

# ---- 2. 构建全部平台产物 ----
echo
echo "[2/5] 构建全部平台产物 (make dist + installer-windows)..."
make dist
make installer-windows

# 收集产物路径
ASSETS=()
while IFS= read -r line; do
  [ -n "$line" ] && ASSETS+=("$line")
done < <(find dist -maxdepth 2 -type f \( \
  -name "${BINARY}-${NEW}-windows-setup.exe" -o \
  -name "${BINARY}-${NEW}-windows.zip" -o \
  -name "${BINARY}-${NEW}-macos.zip" -o \
  -name "${BINARY}-${NEW}-linux.zip" \))
if [ "${#ASSETS[@]}" -eq 0 ]; then
  echo "❌ 未找到任何带版本号 $NEW 的产物, 构建可能失败" >&2
  exit 1
fi
echo "产物清单:"
for a in "${ASSETS[@]}"; do echo "  - $a"; done

# ---- 3. 创建 tag 并推送双远端 ----
echo
echo "[3/5] 创建 tag v${NEW} 并推送到 gitee + origin..."
git tag -a "v${NEW}" -m "v${NEW}" 2>/dev/null || echo "⚠ tag v${NEW} 已存在, 直接推送..."
git push gitee "v${NEW}"
git push origin "v${NEW}"

# ---- 4. 发布 GitHub release ----
echo
echo "[4/5] 发布 GitHub release v${NEW}..."
RELEASE_NOTES="$(build_release_notes "$NEW")"
# 将说明写入临时文件，便于 gh 读取多行文本
NOTES_FILE="$(mktemp)"
printf '%s\n' "$RELEASE_NOTES" > "$NOTES_FILE"
gh release create "v${NEW}" "${ASSETS[@]}" \
  --title "PhiloFTP v${NEW}" \
  --notes-file "$NOTES_FILE" \
  --repo "$GH_REPO" || echo "⚠ GitHub release 创建失败(可能已存在), 尝试上传附件..."
gh release upload "v${NEW}" "${ASSETS[@]}" --clobber --repo "$GH_REPO" || true
rm -f "$NOTES_FILE"

# ---- 5. 发布 Gitee release ----
echo
echo "[5/5] 发布 Gitee release v${NEW}..."
TOKEN="$(gitee_token)"
if [ -z "$TOKEN" ]; then
  echo "❌ 无法从 Keychain 获取 Gitee token, 跳过 Gitee 发布" >&2
else
  # 先尝试创建; 若已存在则取 id
  # 发布说明复用 CHANGELOG 当前版本条目；Gitee API 的 body 需转义 JSON 特殊字符
  BODY="$(changelog_notes "$NEW")"
  BODY_JSON="$(printf '%s' "$BODY" | python3 -c "import sys,json;print(json.dumps(sys.stdin.read()))")"
  CREATE_RESP="$(curl -s --noproxy '*' -X POST "${GITEE_API}/releases?access_token=${TOKEN}" \
    -H "Content-Type: application/json" \
    -d "{\"tag_name\":\"v${NEW}\",\"name\":\"PhiloFTP v${NEW}\",\"body\":${BODY_JSON},\"target_commitish\":\"master\",\"prerelease\":false}")"
  RID="$(echo "$CREATE_RESP" | python3 -c "import sys,json;d=json.load(sys.stdin);print(d.get('id',''))" 2>/dev/null || true)"
  if [ -z "$RID" ]; then
    RID="$(curl -s --noproxy '*' "${GITEE_API}/releases?access_token=${TOKEN}&per_page=20" \
      | python3 -c "import sys,json;d=json.load(sys.stdin);print([r['id'] for r in d if r['tag_name']=='v${NEW}'][0])" 2>/dev/null || true)"
  fi
  if [ -z "$RID" ]; then
    echo "❌ 无法获取 Gitee release id" >&2
  else
    echo "Gitee release id=$RID"
    for f in "${ASSETS[@]}"; do
      echo "  >>> upload $(basename "$f")"
      curl -s --noproxy '*' -X POST "${GITEE_API}/releases/${RID}/attach_files?access_token=${TOKEN}" \
        -F "file=@${f}" > /dev/null
    done
  fi
fi

echo
echo "================================================"
echo " ✅ 发布完成: v${NEW}"
echo "   GitHub: https://github.com/${GH_REPO}/releases/tag/v${NEW}"
echo "   Gitee : https://gitee.com/${GITEE_REPO}/releases/tag/v${NEW}"
echo "================================================"
