import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/viewmodels/auth_viewmodel.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});
  @override State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _form = GlobalKey<FormState>();
  bool _obscure = true;

  @override Widget build(BuildContext context) {
    return Scaffold(backgroundColor: AppColors.white, body: SafeArea(child: Center(child: SingleChildScrollView(
      padding: const EdgeInsets.symmetric(horizontal: 32),
      child: Form(key: _form, child: Consumer<AuthViewModel>(builder: (_, vm, __) => Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const Icon(Icons.fastfood_rounded, size: 56, color: AppColors.primary),
          const SizedBox(height: 12),
          const Text('QuickEats', textAlign: TextAlign.center, style: TextStyle(fontSize: 28, fontWeight: FontWeight.w700)),
          const SizedBox(height: 4),
          const Text('Order ahead, skip the queue', textAlign: TextAlign.center, style: TextStyle(color: AppColors.textSecondary)),
          const SizedBox(height: 40),
          _field(_email, 'Email', Icons.email_outlined, keyboard: TextInputType.emailAddress),
          const SizedBox(height: 16),
          TextFormField(controller: _password, obscureText: _obscure, validator: (v) => v!.isEmpty ? 'Required' : null,
            decoration: InputDecoration(labelText: 'Password', prefixIcon: const Icon(Icons.lock_outline), suffixIcon: IconButton(icon: Icon(_obscure ? Icons.visibility_off : Icons.visibility), onPressed: () => setState(() => _obscure = !_obscure)),
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)), filled: true, fillColor: AppColors.surface)),
          if (vm.error != null) Padding(padding: const EdgeInsets.only(top: 12), child: Text(vm.error!, style: const TextStyle(color: AppColors.error, fontSize: 13))),
          const SizedBox(height: 24),
          SizedBox(height: 50, child: FilledButton(
            onPressed: vm.isLoading ? null : () async { if (_form.currentState!.validate()) { final ok = await vm.login(_email.text.trim(), _password.text); if (ok && mounted) context.go('/shops'); }},
            style: FilledButton.styleFrom(backgroundColor: AppColors.primary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
            child: vm.isLoading ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white)) : const Text('Sign In', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          )),
          const SizedBox(height: 16),
          TextButton(onPressed: () => context.go('/register'), child: const Text("Don't have an account? Register")),
        ],
      ))),
    ))));
  }

  Widget _field(TextEditingController c, String label, IconData icon, {TextInputType? keyboard}) => TextFormField(controller: c, keyboardType: keyboard, validator: (v) => v!.isEmpty ? 'Required' : null,
    decoration: InputDecoration(labelText: label, prefixIcon: Icon(icon), border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)), filled: true, fillColor: AppColors.surface));

  @override void dispose() { _email.dispose(); _password.dispose(); super.dispose(); }
}
