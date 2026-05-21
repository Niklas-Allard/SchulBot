<script setup lang="ts">
import { Link, usePage } from '@inertiajs/vue3';
import { BookOpen, Clock, LayoutGrid, Shield, Wrench } from 'lucide-vue-next';
import { computed } from 'vue';
import AppLogo from '@/components/AppLogo.vue';
import NavFooter from '@/components/NavFooter.vue';
import NavMain from '@/components/NavMain.vue';
import NavUser from '@/components/NavUser.vue';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
} from '@/components/ui/sidebar';
import { dashboard } from '@/routes';
import type { NavItem } from '@/types';

const page = usePage();
const isAdmin = computed(() => (page.props.auth.user as any)?.role === 'admin');

const mainNavItems: NavItem[] = [
    { title: 'Dashboard',    href: dashboard(),      icon: LayoutGrid },
    { title: 'Bot einrichten', href: '/bot/setup',   icon: Wrench },
    { title: 'Verlauf',      href: '/bot/history',   icon: Clock },
];

const adminNavItems = computed<NavItem[]>(() =>
    isAdmin.value ? [{ title: 'Admin', href: '/admin/users', icon: Shield }] : [],
);

const footerNavItems: NavItem[] = [
    { title: 'Einrichtungs-Guide', href: '/guide', icon: BookOpen },
];
</script>

<template>
    <Sidebar collapsible="icon" variant="inset">
        <SidebarHeader>
            <SidebarMenu>
                <SidebarMenuItem>
                    <SidebarMenuButton size="lg" as-child>
                        <Link :href="dashboard()">
                            <AppLogo />
                        </Link>
                    </SidebarMenuButton>
                </SidebarMenuItem>
            </SidebarMenu>
        </SidebarHeader>

        <SidebarContent>
            <NavMain :items="mainNavItems" />
            <NavMain v-if="adminNavItems.length" :items="adminNavItems" />
        </SidebarContent>

        <SidebarFooter>
            <NavFooter :items="footerNavItems" />
            <NavUser />
        </SidebarFooter>
    </Sidebar>
    <slot />
</template>
