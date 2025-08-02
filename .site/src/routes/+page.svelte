<script>
  import { onMount } from 'svelte';
  let commands = [];

  onMount(async () => {
    
        const res = await fetch(`${import.meta.env.BASE_URL}commands.json`);
    commands = await res.json();
  });
</script>

<table>
  <thead>
    <tr>
      <th>Name</th>
      <th>Usage</th>
      <th>Description</th>
      <th>Cooldowns</th>
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
      </tr>
    {/each}
  </tbody>
</table>

