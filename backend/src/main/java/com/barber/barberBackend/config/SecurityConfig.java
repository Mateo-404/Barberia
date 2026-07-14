package com.barber.barberBackend.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import java.net.URI;

import org.springframework.http.HttpMethod;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.http.ProblemDetail;
import org.springframework.security.config.Customizer;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.crypto.argon2.Argon2PasswordEncoder;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

import com.barber.barberBackend.auth.JwtAuthenticationFilter;
import com.barber.barberBackend.auth.JwtService;
import com.fasterxml.jackson.databind.ObjectMapper;

@Configuration
public class SecurityConfig {

    private final JwtService jwtService;
    private final ObjectMapper objectMapper;

    public SecurityConfig(JwtService jwtService, ObjectMapper objectMapper) {
        this.jwtService = jwtService;
        this.objectMapper = objectMapper;
    }

    @Bean
    public Argon2PasswordEncoder passwordEncoder() {
        return new Argon2PasswordEncoder(16, 32, 1, 19456, 2);
    }

    @Bean
    public JwtAuthenticationFilter jwtAuthFilter() {
        return new JwtAuthenticationFilter(jwtService, objectMapper);
    }

    @Bean
    public SecurityFilterChain filterChain(HttpSecurity http) throws Exception {
        http
            .cors(Customizer.withDefaults())
            .csrf(AbstractHttpConfigurer::disable)
            .sessionManagement(sm -> sm.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
            .authorizeHttpRequests(auth -> auth
                .requestMatchers("/ping").permitAll()
                .requestMatchers(HttpMethod.POST, "/turnos").permitAll()
                .requestMatchers(HttpMethod.GET, "/turnos/findDateTimes").permitAll()
                .requestMatchers(HttpMethod.GET, "/servicios", "/servicios/**").permitAll()
                .requestMatchers(HttpMethod.POST, "/administradores/login").permitAll()
                .requestMatchers("/v3/api-docs/**", "/swagger-ui/**").permitAll()
                .anyRequest().authenticated()
            )
            .exceptionHandling(exceptions -> exceptions
                .authenticationEntryPoint((request, response, authException) -> {
                    ProblemDetail pd = ProblemDetail.forStatus(HttpStatus.UNAUTHORIZED);
                    pd.setTitle("No autorizado");
                    pd.setDetail("Se requiere autenticaci\u00f3n para acceder a este recurso");
                    pd.setType(URI.create("/errors/unauthorized"));
                    response.setStatus(HttpStatus.UNAUTHORIZED.value());
                    response.setContentType(MediaType.APPLICATION_PROBLEM_JSON_VALUE);
                    objectMapper.writeValue(response.getWriter(), pd);
                })
            )
            .addFilterBefore(jwtAuthFilter(), UsernamePasswordAuthenticationFilter.class);
        return http.build();
    }
}
