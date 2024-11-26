import org.springframework.cglib.core.Local;

import jakarta.persistence.Basic;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.OneToOne;
import jakarta.persistence.Temporal;
import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

@Entity
@Getter @Setter
@NoArgsConstructor
public class Turno {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Basic
    private long id;
    private boolean cancelado; 
    @Temporal(TemporalType.DATE)
    private LocalDate fecha;
    @OneToOne
    private Cliente cliente;

    // Constructor
    public Turno(long id, Date fecha, Cliente cliente) {
        this.id = id;
        this.fecha = fecha;
        this.cancelado = false;
        this.cliente = cliente;
    }

}
