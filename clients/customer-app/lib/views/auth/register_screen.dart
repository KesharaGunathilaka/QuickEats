import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:quickeats_customer/core/constants/app_colors.dart';
import 'package:quickeats_customer/viewmodels/auth_viewmodel.dart';

class RegisterScreen extends StatefulWidget {
  const RegisterScreen({super.key});
  @override State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final _form = GlobalKey<FormState>();
  final _orgSlug = TextEditingController();
  final _email = TextEditingController();
  final _name = TextEditingController();
  final _password = TextEditingController();

  @override Widget build(BuildContext context) {
    return Scaffold(backgroundColor: AppColors.white, appBar: AppBar(backgroundColor: AppColors.white, elevation: 0),
      body: SingleChildScrollView(padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Form(key: _form, child: Consumer<AuthViewModel>(builder: (_, vm, __) => Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text('Create Account', style: TextStyle(fontSize: 24, fontWeight: FontWeight.w700)),
            const SizedBox(height: 4),
            const Text('Join your university food network', style: TextStyle(color: AppColors.textSecondary)),
            const SizedBox(height: 32),
            _field(_orgSlug, 'Organisation Code', Icons.school_outlined, hint: 'e.g., uor-eng'),
            const SizedBox(height: 16),
            _field(_name, 'Full Name', Icons.person_outlined),
            const SizedBox(height: 16),
            _field(_email, 'University Email', Icons.email_outlined, keyboard: TextInputType.emailAddress),
            const SizedBox(height: 16),
            _field(_password, 'Password', Icons.lock_outline, obscure: true),
            if (vm.error != null) Padding(padding: const EdgeInsets.only(top: 12), child: Text(vm.error!, style: const TextStyle(color: AppColors.error, fontSize: 13))),
            const SizedBox(height: 24),
            SizedBox(height: 50, child: FilledButton(
              onPressed: vm.isLoading ? null : () async {
                if (_form.currentState!.validate()) {
                  final ok = await vm.register(_orgSlug.text.trim(), _email.text.trim(), _password.text, _name.text.trim());
                  if (ok && mounted) { ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('Account created! Please sign in.'))); context.go('/login'); }
                }},
              style: FilledButton.styleFrom(backgroundColor: AppColors.primary, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
              child: vm.isLoading ? const SizedBox(height: 20, width: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white)) : const Text('Register', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            )),
            const SizedBox(height: 16),
            TextButton(onPressed: () => context.go('/login'), child: const Text('Already have an account? Sign In')),
          ],
        ))),
      ),
    );
  }

  Widget _field(TextEditingController c, String label, IconData icon, {String? hint, bool obscure = false, TextInputType? keyboard}) => TextFormField(
    controller: c, obscureText: obscure, keyboardType: keyboard, validator: (v) => v!.isEmpty ? 'Required' : null,
    decoration: InputDecoration(labelText: label, hintText: hint, prefixIcon: Icon(icon), border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)), filled: true, fillColor: AppColors.surface));

  @override void dispose() { _orgSlug.dispose(); _email.dispose(); _name.dispose(); _password.dispose(); super.dispose(); }
}
