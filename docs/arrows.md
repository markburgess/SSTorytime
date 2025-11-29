
# How to use arrows

Getting to grips with notes can feel like a challenge, so don't try to do too much.
It's a simple knowledge management lesson: start with fundamentals and turn them into habits,
then you can build on that.
The important thing is to write things down quickly, before you forget or lose the will.
So it's better to write this:
<pre>

  I saw three ships (blah) it's a song we used to sing
        "           (something) primary school song
        "           (tbd) get the lyrics 
</pre>
than to stop and try to write a perfect documentation, with correct labels.
Later, when you look at it again, you can start changing "blah", "something", and "tbd"
into more useful arrow names that will mean something when you want to search for them later.

For example, later when you are calm, you might want to change this to something like this:
<pre>
  I saw three ships (note) it's a song we used to sing
        "           (first heard) in primary school
        "           (wiki) "https://en.wikipedia.org/wiki/I_Saw_Three_Ships"
        "           (has lyrics) "I saw three ships come sailing in
On Christmas day, on Christmas day
I saw three ships come sailing in
On Christmas day in the morning
And what was in those ships, all three
On Christmas day, on Christmas day?
And what was in those ships, all three
On Christmas day in the morning?
..."

</pre>
To do that, you would have to define the arrows in parentheses. But you can also 
get quite far with just a few that are already defined.

## Starting with a few ...

Try starting with these basic arrows:

* `(then)` - a LEADSTO arrow. You can always join up subsequent events `a (then) b (then) c` etc.
Even if this isn't very specific, you will understand what it means, and you can always go back and change it.

* `(contains)` - a CONTAINS arrows, pretty obvious, though not as common as we might think.

* `(see)` - a NEAR of SIMILAR arrow, useful for just adding a kind of footnote.

Then the most common kind of arrow is EXPRESS PROPERTY, so we add a few that are in common
usage:

* `(e.g.)` - for example, add an example of the thing before the arrow
* `(note)` - add a note about the thing before the arrow
* `(tbd)` - "to be discussed/decided" no idea how to label this, will come back to it! 


This is enough. Now make notes!

<pre>

-some notes

  # basic sequence

  You put your left leg in (then) you put your left leg out (then) you do the hokey cokey and you turn about

  # We can add a note

  $PREV.3 (note) In the US, they say hokey pokey, not hokey cokey

  :: instructions , destructions ::


  +:: _sequence_ ::   # short cut to using (then)
  
You put ONE HAND in    (e.g.) like Napoleon
One hand out           (e.g.) like Oliver Twist
In, out, in out, shake it all about
You do the Hokey Cokey
And you turn around
That’s what it’s all about.

Whoa-o the Hokey Cokey [1st]  (note) how does this change in America? Nothing to do with Cola.
Whoa-o the Hokey Cokey [2nd]
Whoa-o the Hokey Cokey [3rd]
Knees bend, arms stretch rah, rah, rah!  (tbd) this is a kind of yoga for hokey people

  -:: _sequence_ ::  # end auto (then)

Hokey Cokey (see also) Hokey Pokey

</pre>

Then searching try it:

<pre>
\notes "some notes"
</pre>

## Editing your own arrows

To create new arrows, you go to the `SST-config/` sub-directory, either in your current directory
or the one above it. Each type of arrow has its own file. You can add as many lines as you like
in the form:

<pre>
 +  forward arrow meaning (short name) - backward arrow meaning (bwd alias)
</pre>

