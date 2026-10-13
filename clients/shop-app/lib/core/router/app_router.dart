import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:quickeats_shop/core/di/service_locator.dart';
import 'package:quickeats_shop/services/auth_service.dart';
import 'package:quickeats_shop/views/auth/login_screen.dart';
import 'package:quickeats_shop/views/orders/order_queue_screen.dart';
import 'package:quickeats_shop/views/orders/order_detail_screen.dart';
import 'package:quickeats_shop/views/menu/menu_management_screen.dart';
import 'package:quickeats_shop/views/menu/add_menu_item_screen.dart';
import 'package:quickeats_shop/views/shell_screen.dart';

final appRouter = GoRouter(
  initialLocation: '/orders',
  redirect: (context, state) {
    final isAuthenticated = getIt<AuthService>().isAuthenticated;
    final isLoginRoute = state.matchedLocation == '/login';

    if (!isAuthenticated && !isLoginRoute) return '/login';
    if (isAuthenticated && isLoginRoute) return '/orders';
    return null;
  },
  routes: [
    GoRoute(path: '/login', builder: (_, __) => const LoginScreen()),
    ShellRoute(
      builder: (_, __, child) => ShellScreen(child: child),
      routes: [
        GoRoute(path: '/orders', builder: (_, __) => const OrderQueueScreen()),
        GoRoute(
          path: '/orders/:id',
          builder: (_, state) => OrderDetailScreen(orderId: state.pathParameters['id']!),
        ),
        GoRoute(path: '/menu', builder: (_, __) => const MenuManagementScreen()),
        GoRoute(path: '/menu/add', builder: (_, __) => const AddMenuItemScreen()),
      ],
    ),
  ],
);
