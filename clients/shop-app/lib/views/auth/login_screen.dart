import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_shop/core/constants/app_colors.dart';
import 'package:quickeats_shop/viewmodels/auth_viewmodel.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});
  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _obscurePassword = true;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.white,
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Form(
              key: _formKey,
              child: Consumer<AuthViewModel>(
                builder: (context, vm, _) => Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Icon(Icons.storefront_rounded, size: 64, color: AppColors.primary),
                    const SizedBox(height: 16),
                    Text('QuickEats', textAlign: TextAlign.center, style: TextStyle(fontSize: 28, fontWeight: FontWeight.w700, color: AppColors.textPrimary)),
                    const SizedBox(height: 4),
                    Text('Shop Dashboard', textAlign: TextAlign.center, style: TextStyle(fontSize: 14, color: AppColors.textSecondary)),
                    const SizedBox(height: 48),

                    TextFormField(
                      controller: _emailController,
                      keyboardType: TextInputType.emailAddress,
                      decoration: _inputDecoration('Email', Icons.email_outlined),
                      validator: (v) => v == null || v.isEmpty ? 'Enter your email' : null,
                    ),
                    const SizedBox(height: 16),

                    TextFormField(
                      controller: _passwordController,
                      obscureText: _obscurePassword,
                      decoration: _inputDecoration('Password', Icons.lock_outline).copyWith(
                        suffixIcon: IconButton(
                          icon: Icon(_obscurePassword ? Icons.visibility_off : Icons.visibility, color: AppColors.textMuted),
                          onPressed: () => setState(() => _obscurePassword = !_obscurePassword),
                        ),
                      ),
                      validator: (v) => v == null || v.isEmpty ? 'Enter your password' : null,
                    ),
                    const SizedBox(height: 8),

                    if (vm.error != null)
                      Padding(
                        padding: const EdgeInsets.only(bottom: 12),
                        child: Text(vm.error!, style: TextStyle(color: AppColors.error, fontSize: 13)),
                      ),

                    const SizedBox(height: 16),
                    SizedBox(
                      height: 50,
                      child: FilledButton(
                        onPressed: vm.isLoading ? null : () => _handleLogin(vm),
                        style: FilledButton.styleFrom(backgroundColor: AppColors.primary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
                        child: vm.isLoading
                          ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                          : const Text('Sign In', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextButton(
                      onPressed: () { vm.clearError(); context.go('/register'); },
                      child: const Text('New here? Register your shop'),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  InputDecoration _inputDecoration(String label, IconData icon) => InputDecoration(
    labelText: label,
    prefixIcon: Icon(icon, color: AppColors.textMuted),
    border: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: AppColors.border)),
    enabledBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: AppColors.border)),
    focusedBorder: OutlineInputBorder(borderRadius: BorderRadius.circular(12), borderSide: BorderSide(color: AppColors.primary, width: 2)),
    filled: true,
    fillColor: AppColors.surface,
  );

  Future<void> _handleLogin(AuthViewModel vm) async {
    if (!_formKey.currentState!.validate()) return;
    final success = await vm.login(_emailController.text.trim(), _passwordController.text);
    if (success && mounted) context.go('/orders');
  }
}
