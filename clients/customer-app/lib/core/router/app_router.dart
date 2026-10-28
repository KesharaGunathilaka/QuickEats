import 'package:go_router/go_router.dart';
import 'package:quickeats_customer/core/di/service_locator.dart';
import 'package:quickeats_customer/services/auth_service.dart';
import 'package:quickeats_customer/views/auth/login_screen.dart';
import 'package:quickeats_customer/views/auth/register_screen.dart';
import 'package:quickeats_customer/views/shops/shop_list_screen.dart';
import 'package:quickeats_customer/views/menu/menu_screen.dart';
import 'package:quickeats_customer/views/cart/cart_screen.dart';
import 'package:quickeats_customer/views/order/checkout_screen.dart';
import 'package:quickeats_customer/views/order/order_status_screen.dart';
import 'package:quickeats_customer/views/order/order_history_screen.dart';
import 'package:quickeats_customer/views/wallet/wallet_screen.dart';
import 'package:quickeats_customer/views/shell_screen.dart';
import 'package:quickeats_customer/views/settings/settings_screen.dart';

final appRouter = GoRouter(
  initialLocation: '/shops',
  redirect: (context, state) {
    final auth = getIt<AuthService>().isAuthenticated;
    final onAuth = state.matchedLocation == '/login' || state.matchedLocation == '/register';
    if (!auth && !onAuth) return '/login';
    if (auth && onAuth) return '/shops';
    return null;
  },
  routes: [
    GoRoute(path: '/login', builder: (_, __) => const LoginScreen()),
    GoRoute(path: '/register', builder: (_, __) => const RegisterScreen()),
    ShellRoute(
      builder: (_, __, child) => ShellScreen(child: child),
      routes: [
        GoRoute(path: '/shops', builder: (_, __) => const ShopListScreen()),
        GoRoute(path: '/shops/:id/menu', builder: (_, s) => MenuScreen(shopId: s.pathParameters['id']!)),
        GoRoute(path: '/cart', builder: (_, __) => const CartScreen()),
        GoRoute(path: '/checkout', builder: (_, __) => const CheckoutScreen()),
        GoRoute(path: '/orders', builder: (_, __) => const OrderHistoryScreen()),
        GoRoute(path: '/orders/:id', builder: (_, s) => OrderStatusScreen(orderId: s.pathParameters['id']!)),
        GoRoute(path: '/wallet', builder: (_, __) => const WalletScreen()),
        GoRoute(path: '/settings', builder: (_, __) => const SettingsScreen()),
      ],
    ),
  ],
);
