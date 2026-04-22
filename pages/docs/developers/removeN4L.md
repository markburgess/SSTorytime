
# Removing, Replacing, or Editing notes

`removeN4L` deletes one [chapter](concepts/glossary.md#chapter) at a time from the database. See the warning below: it is destructive, `-force` is required, and it is intentionally limited in scope.


## Usage

```
removeN4L -force "chapter name"
```

!!! warning "`-force` is required"
    `removeN4L` will **exit without doing anything** if you omit `-force`. This is
    intentional — deletion is destructive, so the tool asks you to opt in explicitly.
    See the flag check at [`src/removeN4L/removeN4L.go:48-71`](https://github.com/markburgess/SSTorytime/blob/main/src/removeN4L/removeN4L.go#L48-L71).

Eventually you will want to update your notes. Some knowledge is long lived, other knowledge is ephemeral.
Apart from [`reminders'](https://github.com/markburgess/SSTorytime/blob/main/examples/reminders.n4l) you
probably don't want to commit short lived information to a database, but nevertheless we need to update
knowledge as we improve it.

Note: modern SSDs don't like being written to too many times. When using them for databases, they will tend to fail more quickly. The more times you wipe and reload data, the quicker an SSD will fail. My experience is that an SSD lasts about 3 years with normal usage.

## Preferred method

The best and most reliable way to update your notes is to use `N4L -wipe -u *.n4l` to upload all
your notes at the same time. `N4L` takes care of all the work and  makes sure everything is consistent.
However, this takes a long time. There is no easy way around this, because graphs are complicated things
with overlapping threads that need to be made consistent. Trying to remove data and then add it back placcces a
lot of cognitive burden on you the user, so you should try to avoid it. To manage knowledge, you need
to develop a management practice, e.g. updating large data changes once a week. 

## Reminders can be handled specially

Reminders are notes that are placed in time-sensitive contexts, like a calendar, e.g. see the
example [reminders.n4l](https://github.com/markburgess/SSTorytime/blob/main/examples/reminders.n4l):
<pre>
- reminders

  :: Thursday.Hr15 ::

  Get ready for date night! (see also) Suggestions for date night 

</pre>
If you want to update reminders regularly, then place them as the last file of notes in your list:
<pre>

$ N4L -wipe -u file1.n4l ....... reminders.n4l

</pre>
Then you can remove the reminders:
<pre>
$ removeN4L -force reminders.n4l
</pre>
and add them back again without fragmentation:
<pre>
$ N4L -u reminders.n4l
</pre>
Reminders might still overlap with more permanent items from other chapters, but this will minimize the
disruption.

## Exit codes & environment

Exit-code behaviour is literal from the source at [`src/removeN4L/removeN4L.go:48-99`](https://github.com/markburgess/SSTorytime/blob/main/src/removeN4L/removeN4L.go#L48-L99). It is not cleanly normalised; the summary below matches the source exactly.

- **Exit `0`** — `DeleteChapter` was invoked and `main` fell off its end. This covers both
  the success path (`Deleted <chapter>` printed) **and** the SQL-error branch that prints
  `Error running deletechapter function: …` — both paths return `0` because the final
  `return` in `main` is unconditional.
- **Exit `1`** — `-force` was omitted. The tool prints
  `Are you sure you want to remove a chapter? Use -force to confirm.` at
  [`removeN4L.go:63-66`](https://github.com/markburgess/SSTorytime/blob/main/src/removeN4L/removeN4L.go#L63-L66) and stops.
- **Exit `2`** — invalid flag **or** no chapter-name positional argument. `Usage()` is
  called, which itself exits `2` at [`removeN4L.go:79`](https://github.com/markburgess/SSTorytime/blob/main/src/removeN4L/removeN4L.go#L79);
  the `os.Exit(1)` at line 60 is unreachable.
- **Panic (no clean exit code)** — if `sst.DB.Query` returns `err != nil`, `row` is `nil`
  and the unconditional `row.Close()` at
  [`removeN4L.go:97`](https://github.com/markburgess/SSTorytime/blob/main/src/removeN4L/removeN4L.go#L97)
  panics with `runtime error: invalid memory address or nil pointer dereference`
  and a stack trace. This is a known wart; document it so operators are not surprised.

!!! warning "`removeN4L` does not return `-1` on DB error"
    Unlike most tools in this suite, `removeN4L` does not map database failure to
    `-1`. A DB-error branch either falls through to `exit 0` (printing the error
    message) or panics at `row.Close()`. If you script around this tool, do **not**
    rely on a non-zero exit to signal deletion failure — check the printed output.

Environment variables:

- `POSTGRESQL_URI` — overrides the hardcoded DSN in [`pkg/SSTorytime/session.go:41`](https://github.com/markburgess/SSTorytime/blob/main/pkg/SSTorytime/session.go#L41).
- `SST_CONFIG_PATH` — unused at deletion time; not consulted by `removeN4L`.

!!! danger "Destructive operation"
    `removeN4L` calls the `DeleteChapter` PL/pgSQL stored procedure, which drops all nodes,
    links, and page-map rows tagged with the chapter. There is no undo. If you do not have a
    copy of the source `.n4l` files you will have to rebuild the chapter manually. Keep N4L
    sources in version control.
