<script lang="ts">
  import svelteLogo from './assets/svelte.svg'
  import viteLogo from '/vite.svg'

  export const prerender = true;

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
      <td>
        <input type="checkbox" disabled ${command.NoPrefix ? "checked=checked" : ""} class="checkbox" />
      </td>
      <td>
        <input type="checkbox" disabled ${command.CanDisable ? "checked=checked" : ""} class="checkbox" />
      </td>
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
<div class="navbar bg-base-100 shadow-sm">
  <div class="flex-1">
    <a class="btn btn-ghost text-xl">monkebot</a>
  </div>
  <div class="flex gap-2">
    <input type="text" placeholder="Search" class="input input-bordered w-24 md:w-auto" />
  </div>
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

</style>
