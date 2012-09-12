<form action="/b/content/insert_comment">
<input type="hidden" name="content_id" value="{{.content._id}}">
<input type="hidden" name="type" value="{{.content.type}}">
<input type="hidden" name="comment_id" value=""> <!-- Seems pointless when inserting, but background logic needs it. Rethink. -->
<br />
{{if is_stranger}}
	<b>Name</b>:<br />
	<input name="guest_name"><br />
	<br />
{{end}}
<b>Comment</b>:<br />
<textarea name="comment_content" rows="8" cols="53"></textarea>
<input type="submit">
</form>