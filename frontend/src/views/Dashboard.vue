<template>
  <Layout>
    <v-container>
      <v-row justify="center" align="center">
        <v-col cols="12">
          <loading v-if="loading" />

          <div v-else>
            <div class="d-flex justify-space-between align-center">
              <h1>Welcome To Ledger</h1>
              <h1 v-if="loading">Loading...</h1>
              <PlaidLink
                clientName="Ledger"
                env="sandbox"
                :link_token="linkToken"
                :products="['auth', 'transactions', 'balance']"
                :webhook="webhook"
                :onLoad="onLinkLoad"
                :onSuccess="onLinkSuccess"
                :onExit="onLinkExit"
                :onEvent="onLinkEvent"
              >
                <v-btn>Link Account with Plaid</v-btn>
              </PlaidLink>
            </div>
            <hr />
            <v-data-table
              :items="tables.items.items"
              :headers="tables.items.headers"
              :no-data-text="noData"
              :single-expand="true"
              :expanded.sync="expanded"
            >
              <template v-slot:item.action="{ item }">
                <v-btn :to="{path: 'accounts',  params: {itemID: }}"
                  >View Accounts</v-btn
                >
                <v-btn @click="removeItem(item.itemID)"
                  >Remove Institution</v-btn
                >
              </template>
              <template v-slot:item.webhookStatusDatetime="{ item }">
                {{
                  item.webhookStatusDatetime
                    | moment('dddd, MMMM Do YYYY, HH:mm:ss')
                }}
              </template>
            </v-data-table>
          </div>
        </v-col>
      </v-row>
    </v-container>
  </Layout>
</template>

<script>
import PlaidLink from '@/components/PlaidLink';
import Layout from '@/components/layouts/Dashboard';
import Loading from '@/components/Loading';

export default {
  components: {
    PlaidLink,
    Layout,
    Loading,
  },
  data() {
    return {
      tables: {
        items: {
          expanded: [],
          headers: [
            {
              text: 'Record ID',
              value: 'itemID',
            },
            {
              text: 'Institution',
              value: 'institution.name',
            },
            {
              text: 'Status',
              value: 'webhookStatusCodeSent',
            },
            {
              text: 'Last Webhook Update',
              value: 'webhookStatusDatetime',
            },
            {
              text: 'Action',
              value: 'action',
            },
          ],
          items: [],
        },
      },
      linkToken: '',
      loading: true,
      webhook:
        'https://ledger.onetwentyseven.dev/api/external/plaid/v1/webhook',
      items: [],
      headers: [
        {
          text: 'Record ID',
          value: 'itemID',
        },
        {
          text: 'Institution',
          value: 'institution.name',
        },
        {
          text: 'Status',
          value: 'webhookStatusCodeSent',
        },
        {
          text: 'Last Webhook Update',
          value: 'webhookStatusDatetime',
        },
        {
          text: 'Action',
          value: 'action',
        },
      ],
      noData:
        'It appears there are no banking institutions attached to your account. Please register one via Plaid in the top right',
    };
  },
  methods: {
    async onLinkSuccess(_, metadata) {
      await this.$ledger.items().createItem(metadata);
      await this.$ledger.items().items();
    },
    onLinkExit(err) {
      console.log('Exited Link...', err);
    },
    onLinkLoad() {
      console.log('Link is loaded....');
    },
    onLinkEvent(eventName, metadata) {
      console.log(`Recieved Link Event: ${eventName}`, metadata);

      switch (eventName) {
        case 'HANDOFF':
          this.loading = true;
          this.loadItems().then(() => (this.loading = false));
      }
    },
    loadItems() {
      return this.$ledger
        .items()
        .items()
        .then(res => {
          this.tables.items.items = res.data;
        });
    },
    removeItem(itemID) {
      return this.$ledger
        .items()
        .deleteItem(itemID)
        .then(async () => await this.loadItems());
    },
    initializeLink() {
      return this.$ledger
        .plaid()
        .linkToken()
        .then(res => {
          this.linkToken = res.data.token;
        });
    },
  },
  async created() {
    await Promise.all([this.initializeLink(), this.loadItems()]).then(
      () => (this.loading = false)
    );
  },
};
</script>
