function initPersona(persona) {
	if (persona === '') { 
		persona = null; 
	}	
	navigator.id.watch({
		loggedInUser: persona,
		onlogin: function(assertion) {
			$.ajax({ 
				type: 'POST',
				url: '/login/',
				data: {assertion: assertion},
				success: function(res, status, xhr) { window.location.reload(); },
				error: function(xhr, status, err) { navigator.id.logout(); }
			});
		},
		onlogout: function() {
			$.ajax({
				type: 'POST',
				url: '/logout/',
				success: function(res, status, xhr) { window.location.reload(); },
				error: function(xhr, status, err) { console.log('logout error'); }
			});
		}
	});
}

function changeLayout(layout) {
	document.cookie = 'layout=' + layout + '; path=/';
	window.location.reload();
}

function addContact(username) {
	$.ajax({ 
		type: 'POST',
		url: '/contacts/add/' + username + '/',
		success: function(res, status, xhr) { window.location.reload(); },
		error: function(xhr, status, err) { console.log('ajax error: ' + status +  ' | ' + err); }
	});
}

function delContact(username) {
	$.ajax({ 
		type: 'POST',
		url: '/contacts/del/' + username + '/',
		success: function(res, status, xhr) { window.location.reload(); },
		error: function(xhr, status, err) { console.log('ajax error: ' + status +  ' | ' + err); }
	});
}

function addFavorite(username, photoId) {
	$.ajax({ 
		type: 'POST',
		url: '/photos/' + username + '/' + photoId + '/fav/',
		success: function(res, status, xhr) { window.location.reload(); },
		error: function(xhr, status, err) { console.log('ajax error: ' + status +  ' | ' + err); }
	});
}

function delFavorite(username, photoId) {
	$.ajax({ 
		type: 'POST',
		url: '/photos/' + username + '/' + photoId + '/unfav/',
		success: function(res, status, xhr) { window.location.reload(); },
		error: function(xhr, status, err) { console.log('ajax error: ' + status +  ' | ' + err); }
	});
}

function delPhoto(username, photoId) {
	if (confirm('Are you sure you want to delete this photo?')) {
		$.ajax({ 
			type: 'POST',
			url: '/photos/' + username + '/' + photoId + '/del/',
			success: function(res, status, xhr) { window.location.replace('/photos/' + username + '/'); },
			error: function(xhr, status, err) { console.log('ajax error: ' + status +  ' | ' + err); }
		});
	}
}

function addComment(username, photoId) {
	$.ajax({ 
		type: 'POST',
		url: '/photos/' + username + '/' + photoId + '/comment/',
		data: $("#form_add_comment").serialize(),
		success: function(res, status, xhr) { 
			$("#form_add_comment textarea").val("");
			window.location.reload(); 
		},
		error: function(xhr, status, err) { console.log('ajax error: ' + status +  ' | ' + err); }
	});
}

function delComment(username, photoId, commentId) {
	if (confirm('Are you sure you want to delete this comment?')) {
		$.ajax({ 
			type: 'POST',
			url: '/photos/' + username + '/' + photoId + '/delcomment/' + commentId + '/',
			success: function(res, status, xhr) { window.location.reload(); },
			error: function(xhr, status, err) { console.log('ajax error: ' + status +  ' | ' + err); }
		});
	}
}