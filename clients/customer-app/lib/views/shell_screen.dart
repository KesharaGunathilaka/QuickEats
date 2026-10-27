import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/core/di/service_locator.dart';
import 'package:quickeats_customer/services/order_notification_service.dart';
import 'package:quickeats_customer/viewmodels/cart_viewmodel.dart';
import 'package:badges/badges.dart' as badges;

class ShellScreen extends StatefulWidget {
  final Widget child;
  const ShellScreen({super.key, required this.child});

  @override
  State<ShellScreen> createState() => _ShellScreenState();
}

class _ShellScreenState extends State<ShellScreen> {
  @override
  void initState() {
    super.initState();
    // First screen after login: ask once for permission to show order notifications (Android 13+).
    getIt<OrderNotificationService>().requestPermission();
  }

  Widget get child => widget.child;

  @override
  Widget build(BuildContext context) {
    final cartCount = context.watch<CartViewModel>().itemCount;
    return Scaffold(
      body: child,
      bottomNavigationBar: NavigationBar(
        selectedIndex: _idx(context),
        onDestinationSelected: (i) => _tap(context, i),
        backgroundColor: AppColors.white,
        indicatorColor: AppColors.primaryLight,
        destinations: [
          const NavigationDestination(icon: Icon(Icons.storefront_outlined), selectedIcon: Icon(Icons.storefront, color: AppColors.primary), label: 'Shops'),
          NavigationDestination(
            icon: cartCount > 0
              ? badges.Badge(badgeContent: Text('$cartCount', style: const TextStyle(color: Colors.white, fontSize: 10)), child: const Icon(Icons.shopping_cart_outlined))
              : const Icon(Icons.shopping_cart_outlined),
            selectedIcon: const Icon(Icons.shopping_cart, color: AppColors.primary),
            label: 'Cart',
          ),
          const NavigationDestination(icon: Icon(Icons.receipt_long_outlined), selectedIcon: Icon(Icons.receipt_long, color: AppColors.primary), label: 'Orders'),
          const NavigationDestination(icon: Icon(Icons.account_balance_wallet_outlined), selectedIcon: Icon(Icons.account_balance_wallet, color: AppColors.primary), label: 'Wallet'),
        ],
      ),
    );
  }

  int _idx(BuildContext context) {
    final loc = GoRouterState.of(context).matchedLocation;
    if (loc.startsWith('/cart') || loc.startsWith('/checkout')) return 1;
    if (loc.startsWith('/orders')) return 2;
    if (loc.startsWith('/wallet')) return 3;
    return 0;
  }

  void _tap(BuildContext context, int i) {
    switch (i) { case 0: context.go('/shops'); case 1: context.go('/cart'); case 2: context.go('/orders'); case 3: context.go('/wallet'); }
  }
}
