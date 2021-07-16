<template>
  <Layout>
    <v-container fill-height>
      <v-layout align-center justify-center>
        <v-flex xs12 sm8 md4>
          <v-card class="elevation-20">
            <v-toolbar dark>
              <v-toolbar-title> Welcome To Ledger </v-toolbar-title>
            </v-toolbar>
            <v-card-text v-if="!$auth.loading">
              <p>
                This Application leverages Auth0 to handle Authentication. To
                login, please click the button below, this will redirect you off
                to Auth0 where you will be able to login with one of our support
                Providers.
              </p>
              <v-row>
                <v-col lg="12">
                  <v-btn
                    block
                    v-if="!$auth.loading && !$auth.isAuthenticated"
                    @click="$auth.loginWithRedirect()"
                  >
                    Login
                  </v-btn>
                  <v-btn
                    block
                    v-else-if="$auth.isAuthenticated"
                    @click="$auth.logout()"
                  >
                    Logout
                  </v-btn>
                  <loader v-else />
                </v-col>
              </v-row>
            </v-card-text>
            <v-card-text v-else>
              <Loading />
            </v-card-text>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </Layout>
</template>

<script>
import Loading from '@/components/Loading';
import Layout from '@/components/layouts/Login';

export default {
  name: 'Login',
  components: {
    Layout,
    Loading
  },
  methods: {
    login() {
      this.$auth.loginWithRedirect();
    }
  },
  updated() {
    if (this.$auth.isAuthenticated) {
      this.$router.push({ name: 'dashboard' });
    }
  },
  created() {
    if (this.$auth.isAuthenticated) {
      this.$router.push({ name: 'dashboard' });
    }
  }
};
</script>

<style></style>
