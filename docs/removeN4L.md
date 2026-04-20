
# Removing, Replacing, or Editing notes

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
$ removeN4L reminders.n4l
</pre>
and add them back again without fragmentation:
<pre>
$ N4L -u reminders.n4l
</pre>
Reminders might still overlap with more permanent items from other chapters, but this will minimize the
disruption.