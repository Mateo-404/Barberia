package com.barber.barberBackend.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.crypto.argon2.Argon2PasswordEncoder;

@Configuration
public class SecurityConfig {

    @Bean
    public Argon2PasswordEncoder passwordEncoder() {
        // OWASP Password Storage Cheat Sheet (2025/2026) minimum recommendation:
        // m=19456 KiB (19 MiB), t=2 (iterations), p=1 (parallelism)
        // Source: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
        return new Argon2PasswordEncoder(16, 32, 1, 19456, 2);
    }
}
