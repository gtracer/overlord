const Clusters = {
    template: `<div>
 <div class="custom-margin">
    <h1>Clusters</h1>
    <table class="table table-striped table-sm table-hover">
        <thead>
            <tr>
                <th>Name</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="cluster in clusters">
                <td>{{cluster}}</td>
            </tr>
        </tbody>
    </table>
</div></div>`,
    data: function () {
        return {
            clusters: {},
        }
    },
    created: function () {
        this.fetchClusters();
        setInterval(this.fetchClusters, 10000)
    },
    methods: {
        fetchClusters: function () {
            this.$http.get(this.$route.params.userid + "/clusters").then(result => {
                this.clusters = result.body;
            }, error => {
                console.error(error);
            });
        },
    },
}

const Minons = {
    template: `<div>
<h1>Machine Pools</h1>
<button onclick= type="button" class="btn btn-primary float-right mb-4">Download Kubeconfig</button>
<div class="table-responsive">
    <table class="table table-striped table-sm">
        <thead>
            <tr>
                <th>Name</th>
                <th>Role</th>
                <th>Status</th>
                <th>Message</th>
            </tr>
        </thead>
        <tbody>
            <tr v-for="minion in minions">
                <td>{{minion.name}}</td>
                <td>{{minion.role}}</td>
                <td>{{minion.status}}</td>
                <td>{{minion.message}}</td>
            </tr>
        </tbody>
    </table>
</div></div>`,
    data: function () {
        return {
            machinepools: {},
        }
    },
    created: function () {
        this.fetchMinions();
        setInterval(this.fetchMinions, 10000)
    },
    methods: {
        fetchMinions: function () {
            this.$http.get(this.$route.params.userid + "/" + this.$route.params.id + "/minions").then(result => {
                this.minions = result.body;
            }, error => {
                console.error(error);
            });
        },
    },
}

const routes = [
    { path: '/', component: Clusters },
    { path: '/:userid/clusters', component: Clusters },
    { path: '/userid/:id/minions', component: Minons },
]

const router = new VueRouter({
    routes // short for `routes: routes`
})

const app = new Vue({
    router
}).$mount('#app')