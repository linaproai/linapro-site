#!/usr/bin/env bash
# Checks that every Markdown file under apps/lina-site/docs/ (Chinese source)
# has a corresponding translation file in every i18n locale directory.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

DOCS_DIR="$REPO_ROOT/apps/lina-site/docs"
I18N_LOCALES=("en")
I18N_BASE="$REPO_ROOT/apps/lina-site/i18n"

missing_files=()

for locale in "${I18N_LOCALES[@]}"; do
    target_dir="$I18N_BASE/$locale/docusaurus-plugin-content-docs/current"

    echo "Checking locale: $locale (target: ${target_dir#$REPO_ROOT/})"

    if [ ! -d "$target_dir" ]; then
        echo "  ERROR: i18n target directory does not exist: ${target_dir#$REPO_ROOT/}"
        missing_files+=("[$locale] target directory missing: ${target_dir#$REPO_ROOT/}")
        continue
    fi

    while IFS= read -r -d '' src_file; do
        rel="${src_file#$DOCS_DIR/}"
        target_file="$target_dir/$rel"

        if [ ! -f "$target_file" ]; then
            missing_files+=("[$locale] docs/$rel")
        fi
    done < <(find "$DOCS_DIR" -name "*.md" -print0 | sort -z)
done

if [ ${#missing_files[@]} -gt 0 ]; then
    echo ""
    echo "ERROR: The following Chinese docs are missing i18n translations:"
    echo ""
    for entry in "${missing_files[@]}"; do
        echo "  $entry"
    done
    echo ""
    echo "Total missing: ${#missing_files[@]} file(s)"
    echo ""
    echo "Run the linasite-sync-i18n skill or manually create the missing"
    echo "translation files in apps/lina-site/i18n/<locale>/docusaurus-plugin-content-docs/current/"
    exit 1
fi

echo ""
echo "OK: all docs have corresponding i18n translations for locales: ${I18N_LOCALES[*]}"
exit 0
