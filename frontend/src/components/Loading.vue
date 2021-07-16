<template>
  <v-container>
    <v-row>
      <v-col class="text-center">
        <v-progress-circular
          indeterminate
          color="white"
          class="my-5"
        ></v-progress-circular>
        <h1>Loading Page Data, please wait....</h1>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
export default {
  name: 'Loading',
};
</script>
