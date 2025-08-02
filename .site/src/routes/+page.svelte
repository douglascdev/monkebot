<script>
  import { onMount } from 'svelte';
  let commands = [];

  onMount(async () => {
    
    const res = await fetch(`${import.meta.env.BASE_URL}commands.json`);
    commands = await res.json();
  });
</script>

<div class="overflow-x-auto rounded-box border border-base-content/5 bg-base-100">
<table class="table">
  <thead>
    <tr>
      <th>Name</th>
      <th>Usage</th>
      <th>Description</th>
      <th>Channel cooldown</th>
      <th>User cooldown</th>
      <th>No prefix</th>
      <th>Can disable</th>
    </tr>
  </thead>
  <tbody>
    {#each commands as cmd}
      <tr>
        <td>{cmd.Name}</td>
        <td>{cmd.Usage}</td>
        <td>{cmd.Description}</td>
        <td>{cmd.ChannelCooldown}</td>
        <td>{cmd.UserCooldown}</td>
        <td>
          <input type="checkbox" disabled class="checkbox" checked={cmd.NoPrefix} />
        </td>
        <td>
          <input type="checkbox" disabled class="checkbox" checked={cmd.CanDisable} />
        </td>
      </tr>
    {/each}
  </tbody>
</table>
</div>

