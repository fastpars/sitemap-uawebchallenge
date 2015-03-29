$('#MakeSitemap').click(function() {
	$('#result').hide();
	var toSend = {
		homeURL: $('#homeURL').val(),
		levels: $('#levels').val()
	};

	$.post('generateSitemap', toSend, function(d) {
		$('.stats').show();
		var invetval = setInterval(function() {
			refreshSitemapStatistic(d.token, invetval)
		}, 400);
	}, "json");
});

function refreshSitemapStatistic(token, invetval) {
	$.post('sitemapStatistic', {token:token} , function(d) {

		for(var i in d) {
			$('#'+i).text(d[i]);
		}

		if(d.done) {
			$('#result').attr('href', 'sitemap.xml?token='+token).show();
			clearInterval(invetval)
		}
	}, "json").error(function() {
		clearInterval(invetval)
	});
}
