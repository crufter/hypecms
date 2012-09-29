<div class="comments" id="comments">
	<h4>{{.content.comment_count}} comments:</h4>
	<a name="comments"></a>
	<div id="Blog1_comments-block-wrapper">
	{{$con := .content}}
	{{if .content.comments}}
		<dl class="avatar-comment-indent" id="comments-block">
			{{range .content.comments}}
			<dt class="comment-author blog-author">
				<div class="avatar-image-container avatar-stock">
					<span dir="ltr">
						<img src="" width="16" height="16" alt="" title="{{._users_created_by.name}}">
						</a>
					</span>
				</div>
				from 
				<a href="/user/{{._users_created_by.name}}" rel="nofollow">{{._users_created_by.name}}</a>
				<a href="/b/content/delete_comment?type={{$con.type}}&content_id={{$con._id}}&comment_id={{.comment_id}}" style="float:right; border-bottom:1px #000;" class="delete"><img src="/template/icon_delete13.gif"></a>
				<div class="clear"></div>
			</dt>
			<dd class="comment-body" id="Blog1_cmt-2075200508431235064">
				<p>{{.comment_content}}</p>
			</dd>
			<dd class="comment-footer">
				<span class="comment-timestamp">
					{{$created := .created}}
					{{date $created "2006.01.02 15:04:05"}}
					<span class="item-control blog-admin pid-191734713">
						<a class="comment-delete" href="" title="Suprimir comentario">
						<img src="/template/icon_delete13.gif">
						</a>
					</span>
				</span>
			</dd>
			{{end}}
		</dl>
	{{else}}
		No comments yet.
	{{end}}
	</div>
	<p class="comment-footer"></p>
	<div class="comment-form">
		<a name="leave_comment"></a>
		<h4 id="comment-post-message">Comment</h4>
		<p></p>
		<p>Feel free to leave a comment.</p>
		
	</div>
	<p></p>
	<div id="backlinks-container">
		<div id="Blog1_backlinks-container"></div>
	</div>
</div>