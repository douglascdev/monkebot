function loadTableData(data) {
  const tableBody = document.querySelector('#commands-table tbody');
  
  data.forEach(command => {
    const row = document.createElement('tr');

    row.innerHTML = `
      <td>${command.Name}</td>
      <td>${command.Aliases.join(', ') || 'None'}</td>
      <td>${command.Usage}</td>
      <td>${command.Description}</td>
      <td>${command.ChannelCooldown}</td>
      <td>${command.UserCooldown}</td>
      <td>${command.NoPrefix}</td>
      <td>${command.CanDisable}</td>
    `;

    tableBody.appendChild(row);
  });
}

let url = 'commands.json';
fetch(url)
.then(res => res.json())
.then(out => loadTableData(out))
.catch(err => console.log(err));
