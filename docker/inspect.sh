# Read-only views of the Docker environment (POSIX sh; sourced by
# incontainer.sh after core.sh). Nothing here changes data.
#
#   inspect                   the whole environment on one screen
#   inspect duplicates        nodes registered more than once
#   inspect search "<term>"   where a search finds (or loses) its matches
#
# Uses the adapters sql, ws, webcurl from incontainer.sh.

# Empty nodes at (0,0): internal placeholders of the importer, not user data.
NOT_PLACEHOLDER="NOT (s = '' AND (nptr).chan = 0 AND (nptr).cptr = 0)"

sql_quote() { printf '%s' "$1" | sed "s/'/''/g"; }

# ---------------------------------------------------------------- overview
inspect_overview() {
    echo
    echo "Database"
    if [ "$(sql "select to_regclass('sst_install_test.state') is not null")" != "t" ]; then
        echo "  initialized   no install marker (run start)"
    else
        fx=$(sql "select value||' '||to_char(updated_at,'YYYY-MM-DD HH24:MI') from sst_install_test.state where key='fixture'")
        bl=$(sql "select to_char(updated_at,'YYYY-MM-DD HH24:MI') from sst_install_test.state where key='baseline' and value='verified'")
        echo "  initialized   ${fx:-unknown} (doors.n4l)${bl:+ · baseline verified $bl}"
    fi
    if [ "$(sql "select to_regclass('public.node') is not null")" != "t" ]; then
        echo "  graph         empty (no node table)"
    else
        nodes=$(sql "select count(*) from node where $NOT_PLACEHOLDER")
        nchap=$(sql "select count(distinct chap) from node where $NOT_PLACEHOLDER and chap <> ''")
        ph=$(sql "select count(*) from node where not ($NOT_PLACEHOLDER)")
        printf '  graph         %s nodes in %s chapters%s\n' "$nodes" "$nchap" "$([ "$ph" -gt 0 ] && echo "  (+$ph empty placeholder nodes)")"
        fixchap=$(sql "select chap from node where s='door' limit 1")
        sql "select count(*), chap from node where $NOT_PLACEHOLDER and chap <> '' group by chap order by 1 desc, 2 limit 15" |
        while IFS='|' read -r n c; do
            if [ -n "$fixchap" ] && [ "$c" = "$fixchap" ]; then mark="   <- needed by verify"; else mark=""; fi
            printf '                %7s  %s%s\n' "$n" "$c" "$mark"
        done
        [ "$nchap" -le 15 ] || echo "                    ...  ($((nchap - 15)) more)"
        dups=$(sql "select count(*) from (select s from node where $NOT_PLACEHOLDER group by s having count(*) > 1) d")
        if [ "$dups" = 0 ]; then echo "  duplicates    none"
        elif [ "$dups" = 1 ]; then echo "  duplicates    1 node text registered more than once   (details: inspect duplicates)"
        else echo "  duplicates    $dups node texts registered more than once   (details: inspect duplicates)"; fi
    fi

    echo
    echo "Tools"
    present=0; missing=""
    for t in $SST_TOOLS; do
        if ws test -x "/work/bin/$t"; then present=$((present + 1)); else missing="$missing $t"; fi
    done
    total=$(echo $SST_TOOLS | wc -w)
    # Approximation: /work/bin/.staging is recreated at the start of every build.
    newer=$(ws sh -c '[ -d /work/bin/.staging ] || exit 0
        cd /src && find pkg cmd go.mod go.sum -type f \( -name "*.go" -o -name go.mod -o -name go.sum -o -path "cmd/server/public/*" \) \
             -newer /work/bin/.staging 2>/dev/null | grep -v "^cmd/demo_pocs/" | head -n 3' 2>/dev/null)
    if [ -n "$missing" ]; then
        echo "  $present of $total built · missing:$missing   -> run start or rebuild"
    elif [ -n "$newer" ]; then
        echo "  $present of $total built · source changed since the last build (e.g. $(echo $newer | cut -d' ' -f1))   -> run rebuild"
    else
        echo "  $present of $total built · source unchanged since the last build   -> up to date (by timestamps)"
    fi

    echo
    echo "Web"
    port=${SST_WEB_PORT:-8443}
    if web_up; then
        if [ "$(web_state)" = "$(web_want)" ]; then state="running the current http_server and certificate"
        else state="reloading (new binary or certificate)"; fi
        echo "  https://localhost:$port · answering · $state"
    else
        echo "  https://localhost:$port · NOT answering   -> run start; see logs web"
    fi
    mode=$(ws readlink /work/tls/current 2>/dev/null || true)
    if [ -n "$mode" ]; then
        until=$(ws openssl x509 -in /work/tls/current/cert.pem -noout -enddate 2>/dev/null | cut -d= -f2)
        case "$mode" in
            mkcert*) issuer=$(ws openssl x509 -in /work/tls/current/cert.pem -noout -issuer 2>/dev/null | sed 's/.*CN *= *//')
                     echo "  certificate   mkcert, CA \"$issuer\", valid until $until   -> no browser warning where mkcert -install ran" ;;
            *)       echo "  certificate   self-signed, valid until $until   -> browsers warn (see docker/tls/README.md)" ;;
        esac
    fi
    echo
}

# ---------------------------------------------------------------- duplicates
inspect_duplicates() {
    if [ "$(sql "select to_regclass('public.node') is not null")" != "t" ]; then
        echo "The graph is empty."; return 0
    fi
    n=$(sql "select count(*) from (select s from node where $NOT_PLACEHOLDER group by s having count(*) > 1) d")
    ph=$(sql "select count(*) from node where not ($NOT_PLACEHOLDER)")
    echo
    if [ "$n" = 0 ]; then
        echo "Nodes registered more than once: none."
    else
        echo "Nodes registered more than once: $n"
        echo
        sql "select count(*), s, string_agg(distinct coalesce(nullif(chap,''),'(no chapter)'), ' · ')
             from node where $NOT_PLACEHOLDER group by s having count(*) > 1 order by 1 desc, 2 limit 30" |
        while IFS='|' read -r c s ch; do
            printf '  %3sx  %-30s in chapters: %s\n' "$c" "$s" "$ch"
        done
        [ "$n" -le 30 ] || echo "  ... ($((n - 30)) more)"
        echo
        echo "Cause:  usually a chapter added with N4L -u (without -wipe) that reuses existing nodes."
        echo "        The same text in different chapters can also be intended; check the list."
        echo "Fix:    rebuild with all files in one N4L -u -wipe call  (docker/README.md -> Loading data)"
    fi
    [ "$ph" = 0 ] || echo "($ph empty placeholder nodes (0,0) not counted)"
    echo
}

# ---------------------------------------------------------------- search
inspect_search() {
    term=$1
    [ -n "$term" ] || { echo "usage: inspect search \"<term>\"   e.g. inspect search door   or   inspect search \"(zai chengshi)\"" >&2; return 2; }
    case "$term" in
        \(*\)) bare=${term#\(}; bare=${bare%\)}; mode="parentheses = accent-insensitive search" ;;
        *)     bare=$term; mode="plain search (accents must match)" ;;
    esac
    q=$(sql_quote "$bare")
    echo
    echo "Term \"$term\"  -> $mode"
    echo

    if [ "$(sql "select to_regclass('public.node') is not null")" != "t" ]; then
        echo "  the graph is empty"; echo; return 0
    fi
    case "$mode" in
        parentheses*)
            like="UnCmp(s,true) LIKE lower('%$q%')"; label1="database, accent-insensitive"
            fts="unsearch @@ plainto_tsquery('english', '$q')";  label2="database, full-text unaccented" ;;
        *)
            like="lower(s) LIKE lower('%$q%')";      label1="database, text contains"
            fts="search @@ plainto_tsquery('english', '$q')";    label2="database, full-text" ;;
    esac
    n1=$(sql "select count(*) from node where $NOT_PLACEHOLDER and $like" 2>/dev/null || echo "?")
    samples=$(sql "select s from node where $NOT_PLACEHOLDER and $like order by length(s), s limit 2" 2>/dev/null | paste -sd'|' - | sed 's/|/ · /g')
    n2=$(sql "select count(*) from node where $NOT_PLACEHOLDER and $fts" 2>/dev/null || echo "?")
    printf '  %-34s %5s   %s\n' "$label1" "$n1" "$samples"
    printf '  %-34s %5s\n'      "$label2" "$n2"

    cli=$(ws searchN4L "$term" 2>&1 || true)
    n3=$(printf '%s\n' "$cli" | grep -cE '^[0-9]+[:.] ' || true)
    intent=$(printf '%s\n' "$cli" | sed -n 's/^ *Intentional *: *//p' | head -n 1)
    note=""
    [ -z "$intent" ] && [ "$n3" = 0 ] && note="   <- term arrives empty (\"Intentional:\" is blank)"
    printf '  %-34s %5s%s\n' "searchN4L (command line)" "$n3" "$note"

    nweb=""
    if web_up; then
        resp=$(webcurl -sS -w '\n%{http_code}' --data-urlencode "name=$term" https://localhost:8443/searchN4L 2>&1 || true)
        code=$(printf '%s\n' "$resp" | tail -n 1)
        body=$(printf '%s\n' "$resp" | sed '$d')
        # Needles: the term itself plus the texts the database matched (results
        # show stored, possibly accented text, e.g. "zai" finds "zài").
        needles=$(printf '%s\n%s\n' "$bare" "$(sql "select s from node where $NOT_PLACEHOLDER and $like order by length(s), s limit 5" 2>/dev/null)" | sed '/^$/d')
        info_w=$(printf '%s' "$body" | ws jq -r --arg n "$needles" '
            ($n | ascii_downcase | split("\n")) as $ns
            | "\(.Response // "?")|\([.. | strings | ascii_downcase | select(. as $v | any($ns[]; . as $x | $v | contains($x)))] | length)"' 2>/dev/null || echo "?|?")
        nweb=${info_w#*|}
        printf '  %-34s %5s   HTTP %s, Response=%s (values containing the term or a database match)\n' "web search (/searchN4L)" "$nweb" "$code" "${info_w%%|*}"
    else
        printf '  %-34s %5s\n' "web search (/searchN4L)" "-"
    fi

    echo
    if [ "$n1" = 0 ] && [ "$n2" = 0 ]; then
        echo "Diagnosis: no matching data in the database. Check spelling and accents, or load the data."
    elif [ "$n3" = 0 ] && [ "$nweb" = 0 ]; then
        echo "Diagnosis: the data is there and the database finds it, but both the command-line and the"
        echo "           web search return nothing: the term gets lost in the search -> application, not your data."
    elif [ "$n3" = 0 ]; then
        echo "Diagnosis: the data is there and the database finds it, but the command-line search returns"
        echo "           nothing: the search tool loses the term -> application, not your data."
    else
        echo "Diagnosis: found in the database and by the command-line search."
    fi
    echo "(read-only; the counts use different matching rules, so they need not be equal)"
    echo
}
