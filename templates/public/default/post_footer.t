<div class="post-footer">
	<div class="post-footer-line post-footer-line-1">
		{{if .comments}}
		<!--<span class="date-header">time stamp...</span>-->
		<span class="post-comment-link"><a class="comment-link" href="/{{.slug}}#comments">{{$comment_count}} comments</a></span>
		{{else}}
		<span class="post-comment-link"><a class="comment-link" href="/{{.slug}}#leave_comment">Leave a comment</a></span>
		{{end}}
		<span class="post-icons">
			<span class="item-control blog-admin pid-191734713">
				<a href="" title="Editar entrada">
				<img alt="" class="icon-action" height="18" src="icon18_edit_allbkg.gif" width="18" /></a>
			</span>
		</span>
	</div>						
	<div class="post-footer-line post-footer-line-2"></div>
	<div class="post-footer-line post-footer-line-3"></div>
</div>