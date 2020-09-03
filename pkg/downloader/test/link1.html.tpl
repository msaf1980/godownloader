<html>
  <head>
    <meta charset="utf-8">
    <title>Link 1</title>
    <link rel="stylesheet" href="style.css" tppabs="{{ Host }}/style.css">
  </head>
  <body>
    <h1>Link 1</h1>
    <img src="1.gif" tppabs="{{ Host }}/1.gif" /><br>
    <a href="1.gz" tppabs="{{ Host }}/1.gz">Archive</a><br>
    <a href="#1">#1</a><br>
    <a href="#2">#2</a><br>
    <a href="link2.html#3" tppabs="{{ Host }}/link2.html#3">Link 2#3</a><br>
    <a href="not_found.html#1" tppabs="{{ Host }}/not_found.html#1">Link not found</a><br>
    <a href="index.html" tppabs="{{ Host }}/index.html">Up</a><br>
  </body>
</html>
