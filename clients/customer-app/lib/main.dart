import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/core/di/service_locator.dart';
import 'package:quickeats_customer/core/router/app_router.dart';
import 'package:quickeats_customer/services/auth_service.dart';
import 'package:quickeats_customer/viewmodels/auth_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/shops_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/cart_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/order_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/card_viewmodel.dart';
import 'package:quickeats_customer/viewmodels/settings_viewmodel.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await setupServiceLocator();
  runApp(const QuickEatsApp());
}

class QuickEatsApp extends StatelessWidget {
  const QuickEatsApp({super.key});
  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: getIt<AuthService>()),
        ChangeNotifierProvider(create: (_) => getIt<AuthViewModel>()),
        ChangeNotifierProvider(create: (_) => getIt<ShopsViewModel>()),
        ChangeNotifierProvider.value(value: getIt<CartViewModel>()),
        ChangeNotifierProvider(create: (_) => getIt<OrderViewModel>()),
        ChangeNotifierProvider(create: (_) => getIt<CardViewModel>()),
        ChangeNotifierProvider(create: (_) => getIt<SettingsViewModel>()),
      ],
      child: MaterialApp.router(
        title: 'QuickEats',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(colorSchemeSeed: AppColors.primary, useMaterial3: true, scaffoldBackgroundColor: AppColors.surface),
        routerConfig: appRouter,
      ),
    );
  }
}
