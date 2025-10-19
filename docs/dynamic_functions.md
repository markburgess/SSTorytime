
# Dynamic functions in graph node text

When using a knowledge base as a link to the real present world, we can't expect to use only the
static snapshot of data in the database as a source of truth. A useful feature it to be able
to post-process what the database serves, in order to expand variable content on the fly.

**This is a new feature, with potential to be developed in future. Functions must be read-only for
data security, or operate in a sandbox.**

## Examples

Currently, there are only two example functions, demonstrated in the [reminders.n4l](../examples/reminders.n4l)
example. 
<pre>
 Dynamic: Time remaining until Christmas ... {TimeUntil 25 December}

 Dynamic: Regular coordination meeting at 11:30 - TIME REMAINING .. {TimeUntil Hr11 Min30} !!

 Dynamic: {TimeSince Day25 May Yr2018 Hr13} have elapsed since the ChiTek-i company was founded

</pre>
Dynamic content should start with the string `Dynamic: ` and may contain functions which are
enclosed in braces `{function_name arguments}`.

- **TimeUntil** calculates the time until the specificed time.
- **TimeSince** calculates the time since a specified time.

Times are represented using the same class names as one uses for `:: tag ... ::` content.