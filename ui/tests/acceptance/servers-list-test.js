import { currentURL } from '@ember/test-helpers';
import { module, test } from 'qunit';
import { setupApplicationTest } from 'ember-qunit';
import setupMirage from 'ember-cli-mirage/test-support/setup-mirage';
import { findLeader } from '../../mirage/config';
import ServersList from 'nomad-ui/tests/pages/servers/list';

const minimumSetup = () => {
  server.createList('node', 1);
  server.createList('agent', 1);
};

const agentSort = leader => (a, b) => {
  if (`${a.address}:${a.tags.port}` === leader) {
    return 1;
  } else if (`${b.address}:${b.tags.port}` === leader) {
    return -1;
  }
  return 0;
};

module('Acceptance | servers list', function(hooks) {
  setupApplicationTest(hooks);
  setupMirage(hooks);

  test('/servers should list all servers', function(assert) {
    server.createList('node', 1);
    server.createList('agent', 10);

    const leader = findLeader(server.schema);
    const sortedAgents = server.db.agents.sort(agentSort(leader)).reverse();

    ServersList.visit();

    assert.equal(ServersList.servers.length, ServersList.pageSize, 'List is stopped at pageSize');

    ServersList.servers.forEach((server, index) => {
      assert.equal(server.name, sortedAgents[index].name, 'Servers are ordered');
    });
  });

  test('each server should show high-level info of the server', function(assert) {
    minimumSetup();
    const agent = server.db.agents[0];

    ServersList.visit();

    const agentRow = ServersList.servers.objectAt(0);

    assert.equal(agentRow.name, agent.name, 'Name');
    assert.equal(agentRow.status, agent.status, 'Status');
    assert.equal(agentRow.leader, 'True', 'Leader?');
    assert.equal(agentRow.address, agent.address, 'Address');
    assert.equal(agentRow.serfPort, agent.serf_port, 'Serf Port');
    assert.equal(agentRow.datacenter, agent.tags.dc, 'Datacenter');
  });

  test('each server should link to the server detail page', function(assert) {
    minimumSetup();
    const agent = server.db.agents[0];

    ServersList.visit();

    ServersList.servers.objectAt(0).clickRow();

    assert.equal(currentURL(), `/servers/${agent.name}`, 'Now at the server detail page');
  });

  test('when accessing servers is forbidden, show a message with a link to the tokens page', function(assert) {
    server.create('agent');
    server.pretender.get('/v1/agent/members', () => [403, {}, null]);

    ServersList.visit();

    assert.equal(ServersList.error.title, 'Not Authorized');

    ServersList.error.seekHelp();

    assert.equal(currentURL(), '/settings/tokens');
  });
});
