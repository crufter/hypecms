<form action="/b/content/insert_comment">
<input type="hidden" name="content_id" value="{{.content._id}}">
<input type="hidden" name="type" value="{{.content.type}}">
<input type="hidden" name="comment_id" value=""> <!-- Seems pointless, but background logic needs it. Rethink. -->
<textarea name="comment_content" rows="8" cols="53"></textarea>
<input type="submit">
</form>