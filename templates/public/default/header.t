<!DOCTYPE html>

<html xmlns="http://www.w3.org/1999/xhtml">
<head>
	<meta http-equiv="Content-Type" content="text/html; charset=us-ascii" />
<!--[if IE]> <script> (function() { var html5 = ("abbr,article,aside,audio,canvas,datalist,details," + "figure,footer,header,hgroup,mark,menu,meter,nav,output," + "progress,section,time,video").split(','); for (var i = 0; i < html5.length; i++) { document.createElement(html5[i]); } try { document.execCommand('BackgroundImageCache', false, true); } catch(e) {} })(); </script> <![endif]-->
	<title>HypeCMS</title>
	<link type="text/css" rel="stylesheet" href="/template/widget_css_bundle.css" />
	<link type="text/css" rel="stylesheet" href="/template/authorization.css" />
	<link type="text/css" rel="stylesheet" href="/template/style.css" />
</head>
<body>
	<div id="outer-wrapper">
		<div id="wrap2">
			<div id="header-wrapper">
				<div id="header-inner1">
					<div class="header section" id="header">
						<div class="widget Header" id="Header1">
							<div id="header-inner">
								<div class="titlewrapper">
									<a href="/"><h1 class="title">hypeCMS</h1></a>
								</div>
								<div class="descriptionwrapper">
									<p class="description">
										<span>Work in progress...</span>
									</p>
								</div>
							</div>
						</div>
					</div>
					<div id="nav">
						<div>
							<div class="section" id="pages">
								<div class="widget PageList" id="PageList1">
									<div>
										<ul>
											<li><a href="/">Home</a></li>
											{{if is_admin}}
												<li class="current"><a href="/admin">Admin</a></li>
											{{end}}
											<li><a href="/about-us">About us</a></li>
											<li><a href="/contact">Contact</a></li>
										</ul>
										<div class="clear"></div>
										<span class="widget-item-control">
											<span class="item-control blog-admin">
												<a class="quickedit" href=""onclick="" target="configPageList1" title="Editar">
													<img alt="" height="18" src="/template/icon18_wrench_allbkg.png" width="18" />
												</a>
											</span>
										</span>
										<div class="clear"></div>
									</div>
								</div>
							</div>
						</div>
					</div>
					<div id="header-image"></div>
					<form action="/content-search" id="quick-search" method="get" name="quick-search">
						<p>
						<label for="qsearch">Search:</label>
						<input class="tbox" id="qsearch" name="search" title="Start typing and hit ENTER" type="text" value="Search..." />
						<!-- <input alt="Search" class="btn" name="search" src="/template/search.gif" title="Search" type="image" /></p> -->
						<input alt="Search" type="submit" id="search-submit" value="" style="background-image: url(/template/search.gif); border: solid 0px #000000; margin-top: 4px; width: 25px; height: 24px;" />
					</form>
				</div>
			</div>