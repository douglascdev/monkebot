<script lang="ts">
  import svelteLogo from './assets/svelte.svg'
  import viteLogo from '/vite.svg'
  import Counter from './lib/Counter.svelte'

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
</script>

<main>
  <div class="card">
    <Counter />
  </div>
  <div class="overflow-x-auto rounded-box border border-base-content/5 bg-base-100">
  <table id="commands-table" class="table">
  <thead>
    <tr>
      <th>Name</th>
      <th>Aliases</th>
      <th>Usage</th>
      <th>Description</th>
      <th>Channel cooldown</th>
      <th>User cooldown</th>
      <th>No prefix</th>
      <th>Can disable</th>
    </tr>
  </thead>
  <tbody>
  </tbody>
  </table>
  </div>

</main>

<style>
  .logo {
    height: 6em;
    padding: 1.5em;
    will-change: filter;
    transition: filter 300ms;
  }
  .logo:hover {
    filter: drop-shadow(0 0 2em #646cffaa);
  }
  .logo.svelte:hover {
    filter: drop-shadow(0 0 2em #ff3e00aa);
  }
  .read-the-docs {
    color: #888;
  }
</style>
