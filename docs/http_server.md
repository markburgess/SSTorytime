
# `http_server` and web

The http server provided is a generic browsing interface. It isn't meant to be the last
word on browsing the graph. In principle, every application might have its own custom
interface. This web page illustrates the Web API and is used to develop our thinking around
graphs.

The web server has a single argument:
<pre>
./http_server -resources /data/directory
</pre>
This is a directory path which serves as a root for any file paths referenced in URLs, e.g.
where images of documents may be cached in order to be accessible from links rendered in the
browser. It may include any kind of MIME type, such as music files, images, documents etc.

