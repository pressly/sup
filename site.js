
var github = new GitHub();

github.get("repos/pressly/sup/tags", function (err, tags) {
  if (!err) {
    var elem = document.getElementById('releases'),
    childElem = document.createElement('strong'),
    latestTag = tags[0];

    elem.setAttribute('href', elem.getAttribute('href') + '/tags/' + latestTag.name);
    elem.innerHTML = 'Latest release'
    childElem.innerHTML = latestTag.name.replace('v','');
    elem.appendChild(childElem);
  }
});
